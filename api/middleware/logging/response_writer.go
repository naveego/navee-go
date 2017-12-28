package logging

import "net/http"

type loggingResponseWriter struct {
	writer        http.ResponseWriter
	StatusCode    int
	ContentLength int64
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.writer.Header()
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.ContentLength += int64(len(data))
	return lrw.writer.Write(data)
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.StatusCode = statusCode
	lrw.writer.WriteHeader(statusCode)
}

func newLoggingResponseWriter(writer http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		writer: writer,
	}
}
