#ifndef Provenencia_Bridging_Header_h
#define Provenencia_Bridging_Header_h

#include <stddef.h>
#include <stdint.h>

// provenencia_call status: 0 success (out is protobuf bytes or empty),
// 1 failure (out is protobuf provenencia.engine.v1.Error), 2 malloc failed.
int provenencia_call(int32_t method, const uint8_t *in, size_t inLen, uint8_t **out, size_t *outLen);
void provenencia_free(void *p);

#endif
