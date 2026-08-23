#ifndef Provenance_Bridging_Header_h
#define Provenance_Bridging_Header_h

#include <stddef.h>
#include <stdint.h>

int provenance_call(int32_t method, uint8_t *in, size_t inLen, uint8_t **out, size_t *outLen);
void provenance_free(void *p);

#endif
