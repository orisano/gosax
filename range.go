//go:build (go1.22 && goexperiment.rangefunc) || go1.23

package gosax

// Events is a range function over all events, implementing [iter.Seq2].
// It yields each event in order (except [EventEOF]) and an error, if any is encountered during reading.
func (r *Reader) Events(yield func(Event, error) bool) {
	for {
		event, err := r.Event()
		if event.Type() == EventEOF {
			return
		}
		if !yield(event, err) {
			return
		}
	}
}
