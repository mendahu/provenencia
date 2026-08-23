#ifndef Provenance_Bridging_Header_h
#define Provenance_Bridging_Header_h

#include <stddef.h>
#include <stdint.h>

// provenance_call status: 0 success (out is protobuf bytes or empty),
// 1 failure (out is UTF-8 err.Error()), 2 malloc failed.
int provenance_call(int32_t method, const uint8_t *in, size_t inLen, uint8_t **out, size_t *outLen);
void provenance_free(void *p);

#endif
