package default_

//
//import (
//	"bytes"
//	"encoding"
//	"encoding/base64"
//	"reflect"
//	"sort"
//	"strconv"
//	"sync"
//	"unicode/utf8"
//)
//
//func Marshal(v interface{}) ([]byte, error) {
//	e := newEncodeState()
//
//	err := e.marshal(v, encOpts{escapeHTML: true})
//	if err != nil {
//		return nil, err
//	}
//	buf := append([]byte(nil), e.Bytes()...)
//
//	e.Reset()
//	encodeStatePool.Put(e)
//
//	return buf, nil
//}
//
//// Marshaler is the interface implemented by types that
//// can marshal themselves into valid JSON.
//type Marshaler interface {
//	MarshalDefault() ([]byte, error)
//}
//
//
//
//type MarshalerError struct {
//	Type reflect.Type
//	Err  error
//}
//
//func (e *MarshalerError) Error() string {
//	return "default: error calling MarshalDefault for type " + e.Type.String() + ": " + e.Err.Error()
//}
//
//var hex = "0123456789abcdef"
//
//// An encodeState encodes JSON into a bytes.Buffer.
//type encodeState struct {
//	bytes.Buffer // accumulated output
//	scratch [64]byte
//}
//
//var encodeStatePool sync.Pool
//
//func newEncodeState() *encodeState {
//	if v := encodeStatePool.Get(); v != nil {
//		e := v.(*encodeState)
//		e.Reset()
//		return e
//	}
//	return new(encodeState)
//}
//
//type encOpts struct {
//	// quoted causes primitive fields to be encoded inside JSON strings.
//	quoted bool
//	// escapeHTML causes '<', '>', and '&' to be escaped in JSON strings.
//	escapeHTML bool
//}
//
//
//func valueEncoder(v reflect.Value) convertFunc {
//	if !v.IsValid() {
//		return invalidValueConverter
//	}
//	return typeEncoder(v.Type())
//}
//
//func typeEncoder(t reflect.Type) convertFunc {
//	if fi, ok := converterCache.Load(t); ok {
//		return fi
//	}
//
//	// To deal with recursive types, populate the map with an
//	// indirect func before we build it. This type waits on the
//	// real func (f) to be ready and then calls it. This indirect
//	// func is only used for recursive types.
//	var (
//		wg sync.WaitGroup
//		f  convertFunc
//	)
//	wg.Add(1)
//	fi, loaded := converterCache.LoadOrStore(t, convertFunc(func(e *convertState, v reflect.Value, opts convOpts) {
//		// wait until f is assigned elsewhere
//		wg.Wait()
//		f(e, v, opts)
//	}))
//	if loaded {
//		return fi
//	}
//
//	// Compute the real encoder and replace the indirect func with it.
//	f = newTypeConverter(t, true)
//	wg.Done()
//	converterCache.Store(t, f)
//	return f
//}
//
//func marshalerEncoder(e *encodeState, v reflect.Value, opts encOpts) {
//	if v.Kind() == reflect.Ptr && v.IsNil() {
//		e.WriteString("null")
//		return
//	}
//	m, ok := v.Interface().(Marshaler)
//	if !ok {
//		e.WriteString("null")
//		return
//	}
//	b, err := m.MarshalDefault()
//	if err == nil {
//		// copy JSON into buffer, checking validity.
//		err = compact(&e.Buffer, b, opts.escapeHTML)
//	}
//	if err != nil {
//		e.error(&MarshalerError{v.Type(), err})
//	}
//}
//
//
//func textMarshalerEncoder(e *encodeState, v reflect.Value, opts encOpts) {
//	if v.Kind() == reflect.Ptr && v.IsNil() {
//		e.WriteString("null")
//		return
//	}
//	m := v.Interface().(encoding.TextMarshaler)
//	b, err := m.MarshalText()
//	if err != nil {
//		e.error(&MarshalerError{v.Type(), err})
//	}
//	e.stringBytes(b, opts.escapeHTML)
//}
//
//func addrTextMarshalerEncoder(e *encodeState, v reflect.Value, opts encOpts) {
//	va := v.Addr()
//	if va.IsNil() {
//		e.WriteString("null")
//		return
//	}
//	m := va.Interface().(encoding.TextMarshaler)
//	b, err := m.MarshalText()
//	if err != nil {
//		e.error(&MarshalerError{v.Type(), err})
//	}
//	e.stringBytes(b, opts.escapeHTML)
//}
//
//func interfaceEncoder(e *convertState, v reflect.Value, opts encOpts) {
//	if v.IsNil() {
//		return
//	}
//	e.reflectValue(v.Elem(), opts)
//}
//
//
//
//type mapEncoder struct {
//	elemConv convertFunc
//}
//
//func (me *mapEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
//	if v.IsNil() {
//		e.WriteString("null")
//		return
//	}
//	e.WriteByte('{')
//
//	// Extract and sort the keys.
//	keys := v.MapKeys()
//	sv := make([]reflectWithString, len(keys))
//	for i, v := range keys {
//		sv[i].v = v
//		if err := sv[i].resolve(); err != nil {
//			e.error(&MarshalerError{v.Type(), err})
//		}
//	}
//	sort.Slice(sv, func(i, j int) bool { return sv[i].s < sv[j].s })
//
//	for i, kv := range sv {
//		if i > 0 {
//			e.WriteByte(',')
//		}
//		e.string(kv.s, opts.escapeHTML)
//		e.WriteByte(':')
//		me.elemConv(e, v.MapIndex(kv.v), opts)
//	}
//	e.WriteByte('}')
//}
//
//func newMapEncoder(t reflect.Type) convertFunc {
//	switch t.Key().Kind() {
//	case reflect.String,
//		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
//		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
//	default:
//		if !t.Key().Implements(textMarshalerType) {
//			return unsupportedTypeConverter
//		}
//	}
//	me := &mapEncoder{typeEncoder(t.Elem())}
//	return me.encode
//}
//
//func encodeByteSlice(e *encodeState, v reflect.Value, _ encOpts) {
//	if v.IsNil() {
//		e.WriteString("null")
//		return
//	}
//	s := v.Bytes()
//	e.WriteByte('"')
//	if len(s) < 1024 {
//		// for small buffers, using Encode directly is much faster.
//		dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
//		base64.StdEncoding.Encode(dst, s)
//		e.Write(dst)
//	} else {
//		// for large buffers, avoid unnecessary extra temporary
//		// buffer space.
//		enc := base64.NewEncoder(base64.StdEncoding, e)
//		enc.Write(s)
//		enc.Close()
//	}
//	e.WriteByte('"')
//}
//
//type ptrEncoder struct {
//	elemConv convertFunc
//}
//
//func (pe *ptrEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
//	if v.IsNil() {
//		e.WriteString("null")
//		return
//	}
//	pe.elemConv(e, v.Elem(), opts)
//}
//
//func newPtrEncoder(t reflect.Type) convertFunc {
//	enc := &ptrEncoder{typeEncoder(t.Elem())}
//	return enc.encode
//}
//
//
//
//type reflectWithString struct {
//	v reflect.Value
//	s string
//}
//
//func (w *reflectWithString) resolve() error {
//	if w.v.Kind() == reflect.String {
//		w.s = w.v.String()
//		return nil
//	}
//	if tm, ok := w.v.Interface().(encoding.TextMarshaler); ok {
//		buf, err := tm.MarshalText()
//		w.s = string(buf)
//		return err
//	}
//	switch w.v.Kind() {
//	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//		w.s = strconv.FormatInt(w.v.Int(), 10)
//		return nil
//	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
//		w.s = strconv.FormatUint(w.v.Uint(), 10)
//		return nil
//	}
//	panic("unexpected map key type")
//}
//
//// NOTE: keep in sync with stringBytes below.
//func (e *convertState) string(s string, escapeHTML bool) {
//	e.WriteByte('"')
//	start := 0
//	for i := 0; i < len(s); {
//		if b := s[i]; b < utf8.RuneSelf {
//			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
//				i++
//				continue
//			}
//			if start < i {
//				e.WriteString(s[start:i])
//			}
//			switch b {
//			case '\\', '"':
//				e.WriteByte('\\')
//				e.WriteByte(b)
//			case '\n':
//				e.WriteByte('\\')
//				e.WriteByte('n')
//			case '\r':
//				e.WriteByte('\\')
//				e.WriteByte('r')
//			case '\t':
//				e.WriteByte('\\')
//				e.WriteByte('t')
//			default:
//				// This encodes bytes < 0x20 except for \t, \n and \r.
//				// If escapeHTML is set, it also escapes <, >, and &
//				// because they can lead to security holes when
//				// user-controlled strings are rendered into JSON
//				// and served to some browsers.
//				e.WriteString(`\u00`)
//				e.WriteByte(hex[b>>4])
//				e.WriteByte(hex[b&0xF])
//			}
//			i++
//			start = i
//			continue
//		}
//		c, size := utf8.DecodeRuneInString(s[i:])
//		if c == utf8.RuneError && size == 1 {
//			if start < i {
//				e.WriteString(s[start:i])
//			}
//			e.WriteString(`\ufffd`)
//			i += size
//			start = i
//			continue
//		}
//		// U+2028 is LINE SEPARATOR.
//		// U+2029 is PARAGRAPH SEPARATOR.
//		// They are both technically valid characters in JSON strings,
//		// but don't work in JSONP, which has to be evaluated as JavaScript,
//		// and can lead to security holes there. It is valid JSON to
//		// escape them, so we do so unconditionally.
//		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
//		if c == '\u2028' || c == '\u2029' {
//			if start < i {
//				e.WriteString(s[start:i])
//			}
//			e.WriteString(`\u202`)
//			e.WriteByte(hex[c&0xF])
//			i += size
//			start = i
//			continue
//		}
//		i += size
//	}
//	if start < len(s) {
//		e.WriteString(s[start:])
//	}
//	e.WriteByte('"')
//}
//
//// NOTE: keep in sync with string above.
//func (e *encodeState) stringBytes(s []byte, escapeHTML bool) {
//	e.WriteByte('"')
//	start := 0
//	for i := 0; i < len(s); {
//		if b := s[i]; b < utf8.RuneSelf {
//			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
//				i++
//				continue
//			}
//			if start < i {
//				e.Write(s[start:i])
//			}
//			switch b {
//			case '\\', '"':
//				e.WriteByte('\\')
//				e.WriteByte(b)
//			case '\n':
//				e.WriteByte('\\')
//				e.WriteByte('n')
//			case '\r':
//				e.WriteByte('\\')
//				e.WriteByte('r')
//			case '\t':
//				e.WriteByte('\\')
//				e.WriteByte('t')
//			default:
//				// This encodes bytes < 0x20 except for \t, \n and \r.
//				// If escapeHTML is set, it also escapes <, >, and &
//				// because they can lead to security holes when
//				// user-controlled strings are rendered into JSON
//				// and served to some browsers.
//				e.WriteString(`\u00`)
//				e.WriteByte(hex[b>>4])
//				e.WriteByte(hex[b&0xF])
//			}
//			i++
//			start = i
//			continue
//		}
//		c, size := utf8.DecodeRune(s[i:])
//		if c == utf8.RuneError && size == 1 {
//			if start < i {
//				e.Write(s[start:i])
//			}
//			e.WriteString(`\ufffd`)
//			i += size
//			start = i
//			continue
//		}
//		// U+2028 is LINE SEPARATOR.
//		// U+2029 is PARAGRAPH SEPARATOR.
//		// They are both technically valid characters in JSON strings,
//		// but don't work in JSONP, which has to be evaluated as JavaScript,
//		// and can lead to security holes there. It is valid JSON to
//		// escape them, so we do so unconditionally.
//		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
//		if c == '\u2028' || c == '\u2029' {
//			if start < i {
//				e.Write(s[start:i])
//			}
//			e.WriteString(`\u202`)
//			e.WriteByte(hex[c&0xF])
//			i += size
//			start = i
//			continue
//		}
//		i += size
//	}
//	if start < len(s) {
//		e.Write(s[start:])
//	}
//	e.WriteByte('"')
//}
//
//
//func fillField(f field) field {
//	f.nameBytes = []byte(f.name)
//	f.equalFold = foldFunc(f.nameBytes)
//	return f
//}
//
//// byIndex sorts field by index sequence.
//type byIndex []field
//
//func (x byIndex) Len() int { return len(x) }
//
//func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
//
//func (x byIndex) Less(i, j int) bool {
//	for k, xik := range x[i].index {
//		if k >= len(x[j].index) {
//			return false
//		}
//		if xik != x[j].index[k] {
//			return xik < x[j].index[k]
//		}
//	}
//	return len(x[i].index) < len(x[j].index)
//}
//
//
//// dominantField looks through the fields, all of which are known to
//// have the same name, to find the single field that dominates the
//// others using Go's embedding rules, modified by the presence of
//// JSON tags. If there are multiple top-level fields, the boolean
//// will be false: This condition is an error in Go and we skip all
//// the fields.
//func dominantField(fields []field) (field, bool) {
//	// The fields are sorted in increasing index-length order, then by presence of tag.
//	// That means that the first field is the dominant one. We need only check
//	// for error cases: two fields at top level, either both tagged or neither tagged.
//	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
//		return field{}, false
//	}
//	return fields[0], true
//}
//
