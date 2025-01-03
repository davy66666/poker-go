// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2010 The Go Authors.  All rights reserved.
// https://github.com/davy66666/poker-go/src/github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package proto

/*
 * Types and routines for supporting protocol buffer extensions.
 */

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

// ErrMissingExtension is the error returned by GetExtension if the named extension is not in the message.
var ErrMissingExtension = errors.New("proto: missing extension")

// ExtensionRange represents a range of message extensions for a protocol buffer.
// Used in code generated by the protocol compiler.
type ExtensionRange struct {
	Start, End int32 // both inclusive
}

// extendableProto is an interface implemented by any protocol buffer generated by the current
// proto compiler that may be extended.
type extendableProto interface {
	Message
	ExtensionRangeArray() []ExtensionRange
	extensionsWrite() map[int32]Extension
	extensionsRead() (map[int32]Extension, sync.Locker)
}

// extendableProtoV1 is an interface implemented by a protocol buffer generated by the previous
// version of the proto compiler that may be extended.
type extendableProtoV1 interface {
	Message
	ExtensionRangeArray() []ExtensionRange
	ExtensionMap() map[int32]Extension
}

// extensionAdapter is a wrapper around extendableProtoV1 that implements extendableProto.
type extensionAdapter struct {
	extendableProtoV1
}

func (e extensionAdapter) extensionsWrite() map[int32]Extension {
	return e.ExtensionMap()
}

func (e extensionAdapter) extensionsRead() (map[int32]Extension, sync.Locker) {
	return e.ExtensionMap(), notLocker{}
}

// notLocker is a sync.Locker whose Lock and Unlock methods are nops.
type notLocker struct{}

func (n notLocker) Lock()   {}
func (n notLocker) Unlock() {}

// extendable returns the extendableProto interface for the given generated proto message.
// If the proto message has the old extension format, it returns a wrapper that implements
// the extendableProto interface.
func extendable(p interface{}) (extendableProto, bool) {
	if ep, ok := p.(extendableProto); ok {
		return ep, ok
	}
	if ep, ok := p.(extendableProtoV1); ok {
		return extensionAdapter{ep}, ok
	}
	return nil, false
}

// XXX_InternalExtensions is an internal representation of proto extensions.
//
// Each generated message struct type embeds an anonymous XXX_InternalExtensions field,
// thus gaining the unexported 'extensions' method, which can be called only from the proto package.
//
// The methods of XXX_InternalExtensions are not concurrency safe in general,
// but calls to logically read-only methods such as has and get may be executed concurrently.
type XXX_InternalExtensions struct {
	// The struct must be indirect so that if a user inadvertently copies a
	// generated message and its embedded XXX_InternalExtensions, they
	// avoid the mayhem of a copied mutex.
	//
	// The mutex serializes all logically read-only operations to p.extensionMap.
	// It is up to the client to ensure that write operations to p.extensionMap are
	// mutually exclusive with other accesses.
	p *struct {
		mu           sync.Mutex
		extensionMap map[int32]Extension
	}
}

// extensionsWrite returns the extension map, creating it on first use.
func (e *XXX_InternalExtensions) extensionsWrite() map[int32]Extension {
	if e.p == nil {
		e.p = new(struct {
			mu           sync.Mutex
			extensionMap map[int32]Extension
		})
		e.p.extensionMap = make(map[int32]Extension)
	}
	return e.p.extensionMap
}

// extensionsRead returns the extensions map for read-only use.  It may be nil.
// The caller must hold the returned mutex's lock when accessing Elements within the map.
func (e *XXX_InternalExtensions) extensionsRead() (map[int32]Extension, sync.Locker) {
	if e.p == nil {
		return nil, nil
	}
	return e.p.extensionMap, &e.p.mu
}

var extendableProtoType = reflect.TypeOf((*extendableProto)(nil)).Elem()
var extendableProtoV1Type = reflect.TypeOf((*extendableProtoV1)(nil)).Elem()

// ExtensionDesc represents an extension specification.
// Used in generated code from the protocol compiler.
type ExtensionDesc struct {
	ExtendedType  Message     // nil pointer to the type that is being extended
	ExtensionType interface{} // nil pointer to the extension type
	Field         int32       // field number
	Name          string      // fully-qualified name of extension, for text formatting
	Tag           string      // protobuf tag style
	Filename      string      // name of the file in which the extension is defined
}

func (ed *ExtensionDesc) repeated() bool {
	t := reflect.TypeOf(ed.ExtensionType)
	return t.Kind() == reflect.Slice && t.Elem().Kind() != reflect.Uint8
}

// Extension represents an extension in a message.
type Extension struct {
	// When an extension is stored in a message using SetExtension
	// only desc and value are set. When the message is marshaled
	// enc will be set to the encoded form of the message.
	//
	// When a message is unmarshaled and contains extensions, each
	// extension will have only enc set. When such an extension is
	// accessed using GetExtension (or GetExtensions) desc and value
	// will be set.
	desc  *ExtensionDesc
	value interface{}
	enc   []byte
}

// SetRawExtension is for testing only.
func SetRawExtension(base Message, id int32, b []byte) {
	epb, ok := extendable(base)
	if !ok {
		return
	}
	extmap := epb.extensionsWrite()
	extmap[id] = Extension{enc: b}
}

// isExtensionField returns true iff the given field number is in an extension range.
func isExtensionField(pb extendableProto, field int32) bool {
	for _, er := range pb.ExtensionRangeArray() {
		if er.Start <= field && field <= er.End {
			return true
		}
	}
	return false
}

// checkExtensionTypes checks that the given extension is valid for pb.
func checkExtensionTypes(pb extendableProto, extension *ExtensionDesc) error {
	var pbi interface{} = pb
	// Check the extended type.
	if ea, ok := pbi.(extensionAdapter); ok {
		pbi = ea.extendableProtoV1
	}
	if a, b := reflect.TypeOf(pbi), reflect.TypeOf(extension.ExtendedType); a != b {
		return errors.New("proto: bad extended type; " + b.String() + " does not extend " + a.String())
	}
	// Check the range.
	if !isExtensionField(pb, extension.Field) {
		return errors.New("proto: bad extension number; not in declared ranges")
	}
	return nil
}

// extPropKey is sufficient to uniquely identify an extension.
type extPropKey struct {
	base  reflect.Type
	field int32
}

var extProp = struct {
	sync.RWMutex
	m map[extPropKey]*Properties
}{
	m: make(map[extPropKey]*Properties),
}

func extensionProperties(ed *ExtensionDesc) *Properties {
	key := extPropKey{base: reflect.TypeOf(ed.ExtendedType), field: ed.Field}

	extProp.RLock()
	if prop, ok := extProp.m[key]; ok {
		extProp.RUnlock()
		return prop
	}
	extProp.RUnlock()

	extProp.Lock()
	defer extProp.Unlock()
	// Check again.
	if prop, ok := extProp.m[key]; ok {
		return prop
	}

	prop := new(Properties)
	prop.Init(reflect.TypeOf(ed.ExtensionType), "unknown_name", ed.Tag, nil)
	extProp.m[key] = prop
	return prop
}

// encode encodes any unmarshaled (unencoded) extensions in e.
func encodeExtensions(e *XXX_InternalExtensions) error {
	m, mu := e.extensionsRead()
	if m == nil {
		return nil // fast path
	}
	mu.Lock()
	defer mu.Unlock()
	return encodeExtensionsMap(m)
}

// encode encodes any unmarshaled (unencoded) extensions in e.
func encodeExtensionsMap(m map[int32]Extension) error {
	for k, e := range m {
		if e.value == nil || e.desc == nil {
			// Extension is only in its encoded form.
			continue
		}

		// We don't skip extensions that have an encoded form set,
		// because the extension value may have been mutated after
		// the last time this function was called.

		et := reflect.TypeOf(e.desc.ExtensionType)
		props := extensionProperties(e.desc)

		p := NewBuffer(nil)
		// If e.value has type T, the encoder expects a *struct{ X T }.
		// Pass a *T with a zero field and hope it all works out.
		x := reflect.New(et)
		x.Elem().Set(reflect.ValueOf(e.value))
		if err := props.enc(p, props, toStructPointer(x)); err != nil {
			return err
		}
		e.enc = p.buf
		m[k] = e
	}
	return nil
}

func extensionsSize(e *XXX_InternalExtensions) (n int) {
	m, mu := e.extensionsRead()
	if m == nil {
		return 0
	}
	mu.Lock()
	defer mu.Unlock()
	return extensionsMapSize(m)
}

func extensionsMapSize(m map[int32]Extension) (n int) {
	for _, e := range m {
		if e.value == nil || e.desc == nil {
			// Extension is only in its encoded form.
			n += len(e.enc)
			continue
		}

		// We don't skip extensions that have an encoded form set,
		// because the extension value may have been mutated after
		// the last time this function was called.

		et := reflect.TypeOf(e.desc.ExtensionType)
		props := extensionProperties(e.desc)

		// If e.value has type T, the encoder expects a *struct{ X T }.
		// Pass a *T with a zero field and hope it all works out.
		x := reflect.New(et)
		x.Elem().Set(reflect.ValueOf(e.value))
		n += props.size(props, toStructPointer(x))
	}
	return
}

// HasExtension returns whether the given extension is present in pb.
func HasExtension(pb Message, extension *ExtensionDesc) bool {
	// TODO: Check types, field numbers, etc.?
	epb, ok := extendable(pb)
	if !ok {
		return false
	}
	extmap, mu := epb.extensionsRead()
	if extmap == nil {
		return false
	}
	mu.Lock()
	_, ok = extmap[extension.Field]
	mu.Unlock()
	return ok
}

// ClearExtension removes the given extension from pb.
func ClearExtension(pb Message, extension *ExtensionDesc) {
	epb, ok := extendable(pb)
	if !ok {
		return
	}
	// TODO: Check types, field numbers, etc.?
	extmap := epb.extensionsWrite()
	delete(extmap, extension.Field)
}

// GetExtension parses and returns the given extension of pb.
// If the extension is not present and has no default value it returns ErrMissingExtension.
func GetExtension(pb Message, extension *ExtensionDesc) (interface{}, error) {
	epb, ok := extendable(pb)
	if !ok {
		return nil, errors.New("proto: not an extendable proto")
	}

	if err := checkExtensionTypes(epb, extension); err != nil {
		return nil, err
	}

	emap, mu := epb.extensionsRead()
	if emap == nil {
		return defaultExtensionValue(extension)
	}
	mu.Lock()
	defer mu.Unlock()
	e, ok := emap[extension.Field]
	if !ok {
		// defaultExtensionValue returns the default value or
		// ErrMissingExtension if there is no default.
		return defaultExtensionValue(extension)
	}

	if e.value != nil {
		// Already decoded. Check the descriptor, though.
		if e.desc != extension {
			// This shouldn't happen. If it does, it means that
			// GetExtension was called twice with two different
			// descriptors with the same field number.
			return nil, errors.New("proto: descriptor conflict")
		}
		return e.value, nil
	}

	v, err := decodeExtension(e.enc, extension)
	if err != nil {
		return nil, err
	}

	// Remember the decoded version and drop the encoded version.
	// That way it is safe to mutate what we return.
	e.value = v
	e.desc = extension
	e.enc = nil
	emap[extension.Field] = e
	return e.value, nil
}

// defaultExtensionValue returns the default value for extension.
// If no default for an extension is defined ErrMissingExtension is returned.
func defaultExtensionValue(extension *ExtensionDesc) (interface{}, error) {
	t := reflect.TypeOf(extension.ExtensionType)
	props := extensionProperties(extension)

	sf, _, err := fieldDefault(t, props)
	if err != nil {
		return nil, err
	}

	if sf == nil || sf.value == nil {
		// There is no default value.
		return nil, ErrMissingExtension
	}

	if t.Kind() != reflect.Ptr {
		// We do not need to return a Ptr, we can directly return sf.value.
		return sf.value, nil
	}

	// We need to return an interface{} that is a pointer to sf.value.
	value := reflect.New(t).Elem()
	value.Set(reflect.New(value.Type().Elem()))
	if sf.kind == reflect.Int32 {
		// We may have an int32 or an enum, but the underlying data is int32.
		// Since we can't set an int32 into a non int32 reflect.value directly
		// set it as a int32.
		value.Elem().SetInt(int64(sf.value.(int32)))
	} else {
		value.Elem().Set(reflect.ValueOf(sf.value))
	}
	return value.Interface(), nil
}

// decodeExtension decodes an extension encoded in b.
func decodeExtension(b []byte, extension *ExtensionDesc) (interface{}, error) {
	o := NewBuffer(b)

	t := reflect.TypeOf(extension.ExtensionType)

	props := extensionProperties(extension)

	// t is a pointer to a struct, pointer to basic type or a slice.
	// Allocate a "field" to store the pointer/slice itself; the
	// pointer/slice will be stored here. We pass
	// the address of this field to props.dec.
	// This passes a zero field and a *t and lets props.dec
	// interpret it as a *struct{ x t }.
	value := reflect.New(t).Elem()

	for {
		// Discard wire type and field number varint. It isn't needed.
		if _, err := o.DecodeVarint(); err != nil {
			return nil, err
		}

		if err := props.dec(o, props, toStructPointer(value.Addr())); err != nil {
			return nil, err
		}

		if o.index >= len(o.buf) {
			break
		}
	}
	return value.Interface(), nil
}

// GetExtensions returns a slice of the extensions present in pb that are also listed in es.
// The returned slice has the same length as es; missing extensions will appear as nil elements.
func GetExtensions(pb Message, es []*ExtensionDesc) (extensions []interface{}, err error) {
	epb, ok := extendable(pb)
	if !ok {
		return nil, errors.New("proto: not an extendable proto")
	}
	extensions = make([]interface{}, len(es))
	for i, e := range es {
		extensions[i], err = GetExtension(epb, e)
		if err == ErrMissingExtension {
			err = nil
		}
		if err != nil {
			return
		}
	}
	return
}

// ExtensionDescs returns a new slice containing pb's extension descriptors, in undefined order.
// For non-registered extensions, ExtensionDescs returns an incomplete descriptor containing
// just the Field field, which defines the extension's field number.
func ExtensionDescs(pb Message) ([]*ExtensionDesc, error) {
	epb, ok := extendable(pb)
	if !ok {
		return nil, fmt.Errorf("proto: %T is not an extendable proto.Message", pb)
	}
	registeredExtensions := RegisteredExtensions(pb)

	emap, mu := epb.extensionsRead()
	if emap == nil {
		return nil, nil
	}
	mu.Lock()
	defer mu.Unlock()
	extensions := make([]*ExtensionDesc, 0, len(emap))
	for extid, e := range emap {
		desc := e.desc
		if desc == nil {
			desc = registeredExtensions[extid]
			if desc == nil {
				desc = &ExtensionDesc{Field: extid}
			}
		}

		extensions = append(extensions, desc)
	}
	return extensions, nil
}

// SetExtension sets the specified extension of pb to the specified value.
func SetExtension(pb Message, extension *ExtensionDesc, value interface{}) error {
	epb, ok := extendable(pb)
	if !ok {
		return errors.New("proto: not an extendable proto")
	}
	if err := checkExtensionTypes(epb, extension); err != nil {
		return err
	}
	typ := reflect.TypeOf(extension.ExtensionType)
	if typ != reflect.TypeOf(value) {
		return errors.New("proto: bad extension value type")
	}
	// nil extension values need to be caught early, because the
	// encoder can't distinguish an ErrNil due to a nil extension
	// from an ErrNil due to a missing field. Extensions are
	// always optional, so the encoder would just swallow the error
	// and drop all the extensions from the encoded message.
	if reflect.ValueOf(value).IsNil() {
		return fmt.Errorf("proto: SetExtension called with nil value of type %T", value)
	}

	extmap := epb.extensionsWrite()
	extmap[extension.Field] = Extension{desc: extension, value: value}
	return nil
}

// ClearAllExtensions clears all extensions from pb.
func ClearAllExtensions(pb Message) {
	epb, ok := extendable(pb)
	if !ok {
		return
	}
	m := epb.extensionsWrite()
	for k := range m {
		delete(m, k)
	}
}

// A global registry of extensions.
// The generated code will register the generated descriptors by calling RegisterExtension.

var extensionMaps = make(map[reflect.Type]map[int32]*ExtensionDesc)

// RegisterExtension is called from the generated code.
func RegisterExtension(desc *ExtensionDesc) {
	st := reflect.TypeOf(desc.ExtendedType).Elem()
	m := extensionMaps[st]
	if m == nil {
		m = make(map[int32]*ExtensionDesc)
		extensionMaps[st] = m
	}
	if _, ok := m[desc.Field]; ok {
		panic("proto: duplicate extension registered: " + st.String() + " " + strconv.Itoa(int(desc.Field)))
	}
	m[desc.Field] = desc
}

// RegisteredExtensions returns a map of the registered extensions of a
// protocol buffer struct, indexed by the extension number.
// The argument pb should be a nil pointer to the struct type.
func RegisteredExtensions(pb Message) map[int32]*ExtensionDesc {
	return extensionMaps[reflect.TypeOf(pb).Elem()]
}
