package eventlog

import "compress/gzip"

type RecorderOpt interface{}

type timeSourceOpt func() int64

// TimeSourceOpt can be used to override the default Time source
// for an interceptor.  This can be useful for changing the
// granularity of the timestamps, or picking some externally
// supplied sync point when trying to synchronize logs.
// The default Time source will timestamp with the Time, in
// milliseconds since the interceptor was created.
func TimeSourceOpt(source func() int64) RecorderOpt {
	return timeSourceOpt(source)
}

type retainRequestDataOpt struct{}

// RetainRequestDataOpt indicates that the full request data should be
// embedded into the logs.  Usually, this option is undesirable since although
// request data is not actually needed to replay a log, the request data
// increases the size of the log substantially and the request data
// may be considered sensitive so is therefore unsuitable for
// debug/service.  However, for debugging application code, sometimes,
// having the complete logs is available, so this option may be set
// to true.
func RetainRequestDataOpt() RecorderOpt {
	return retainRequestDataOpt{}
}

type compressionLevelOpt int

// DefaultCompressionLevel is used for event capture when not overridden.
// In empirical tests, best speed was only a few tenths of a percent
// worse than best compression, but your results may vary.
const DefaultCompressionLevel = gzip.BestSpeed

// CompressionLevelOpt takes any of the compression levels supported
// by the golang standard gzip package.
func CompressionLevelOpt(level int) RecorderOpt {
	return compressionLevelOpt(level)
}

// DefaultBufferSize is the number of unwritten state Events which
// may be held in queue before blocking.
const DefaultBufferSize = 5000

type bufferSizeOpt int

// BufferSizeOpt overrides the default buffer size of the
// interceptor buffer.  Once the buffer overflows, the state
// machine will be blocked from receiving new state Events
// until the buffer has room.
func BufferSizeOpt(size int) RecorderOpt {
	return bufferSizeOpt(size)
}

type fileSplitterOpt func(EventRecord) []EventRecord

func FileSplitterOpt(splitter func(EventRecord) []EventRecord) RecorderOpt {
	return fileSplitterOpt(splitter)
}
