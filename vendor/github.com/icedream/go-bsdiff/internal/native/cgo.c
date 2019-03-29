#include "cgo.h"

#include "bsdiff.h"
#include "bspatch.h"

extern int cgo_write_buffer(int bufferIndex, void* buf, int size);

int cgo_write(struct bsdiff_stream* stream,
	const void* buf, int size) {
	struct buffer_table_index* bufferEntry;
	
	bufferEntry = (struct buffer_table_index*)stream->opaque;

	return cgo_write_buffer(bufferEntry->index, (void*)buf, size);
}

extern int cgo_read_buffer(int bufferIndex, void* buf, int size);

int cgo_read(const struct bspatch_stream* stream,
	void* buf, int size) {
		struct buffer_table_index* bufferEntry;
		
		bufferEntry = (struct buffer_table_index*)stream->opaque;
		
	return cgo_read_buffer(bufferEntry->index, buf, size) ;
}

int bsdiff_cgo(uint8_t* oldptr, int64_t oldsize,
	uint8_t* newptr, int64_t newsize,
	int bufferIndex)
{
	struct bsdiff_stream stream;
	stream.malloc = malloc;
	stream.free = free;
	stream.write = cgo_write;

	struct buffer_table_index bufferEntry;
	bufferEntry.index = bufferIndex;
	stream.opaque = &bufferEntry;

	return bsdiff(oldptr, oldsize, newptr, newsize, &stream);
}

int bspatch_cgo(uint8_t* oldptr, int64_t oldsize,
	uint8_t* newptr, int64_t newsize,
	int bufferIndex)
{
	struct bspatch_stream stream;
	stream.read = cgo_read;
	
	struct buffer_table_index bufferEntry;
	bufferEntry.index = bufferIndex;
	stream.opaque = &bufferEntry;
	
	return bspatch(oldptr, oldsize, newptr, newsize, &stream);
}