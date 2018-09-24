package slice

type Stream struct {
	s interface{}
}

func NewStream() *Stream {
	return &Stream{}
}
func (stream *Stream) WithSlice(s interface{}) *Stream {
	stream.s = s
	return stream
}
func (stream *Stream) Filter(f func(interface{}) bool, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(FilterFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) Map(f func(interface{}) interface{}, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(MapFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) Distinct(f func(interface{}, interface{}) int, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(DistinctFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) Sorted(f func(interface{}, interface{}) int, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(SortedFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) Peek(f func(interface{}), ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(PeekFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) Limit(maxSize int, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(LimitFunc(stream.s, maxSize, ifStringAsRune...))
}

func (stream *Stream) Skip(n int, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(SkipFunc(stream.s, n, ifStringAsRune...))
}

func (stream *Stream) TakeWhile(f func(interface{}) bool, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(TakeWhileFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) TakeUntil(f func(interface{}) bool, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(TakeUntilFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) DropWhile(f func(interface{}) bool, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(DropWhileFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) DropUntil(f func(interface{}) bool, ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(DropUntilFunc(stream.s, f, ifStringAsRune...))
}

func (stream *Stream) ForEach(f func(interface{}), ifStringAsRune ...bool) {
	ForEachFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) ForEachOrdered(f func(interface{}), ifStringAsRune ...bool) {
	ForEachOrderedFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) ToSlice(ifStringAsRune ...bool) interface{} {
	return ToSliceFunc(stream.s, ifStringAsRune...)
}

func (stream *Stream) Reduce(f func(left, right interface{}) interface{}, ifStringAsRune ...bool) interface{} {
	return ReduceFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) Min(f func(interface{}, interface{}) int, ifStringAsRune ...bool) interface{} {
	return MinFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) Max(f func(interface{}, interface{}) int, ifStringAsRune ...bool) interface{} {
	return MaxFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) Count(ifStringAsRune ...bool) int {
	return CountFunc(stream.s, ifStringAsRune...)
}

func (stream *Stream) AnyMatch(f func(interface{}) bool, ifStringAsRune ...bool) bool {
	return AnyMatchFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) AllMatch(f func(interface{}) bool, ifStringAsRune ...bool) bool {
	return AllMatchFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) NoneMatch(f func(interface{}) bool, ifStringAsRune ...bool) bool {
	return NoneMatchFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) FindFirst(f func(interface{}) bool, ifStringAsRune ...bool) interface{} {
	return FindFirstFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) FindAny(f func(interface{}) bool, ifStringAsRune ...bool) interface{} {
	return FindAnyFunc(stream.s, f, ifStringAsRune...)
}

func (stream *Stream) Empty(ifStringAsRune ...bool) interface{} {
	return EmptyFunc(stream.s, ifStringAsRune...)
}

func (stream *Stream) Of(ifStringAsRune ...bool) *Stream {
	return stream.WithSlice(Of(stream.s, ifStringAsRune...))
}

func (stream *Stream) Concat(s2 *Stream) *Stream {
	return stream.WithSlice(ConcatFunc(stream.s, s2.s))
}
