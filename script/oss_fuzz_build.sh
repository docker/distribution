#!/bin/bash -eu

compile_go_fuzzer github.com/distribution/distribution/v3/configuration ParserFuzzer parser_fuzzer
compile_go_fuzzer github.com/distribution/distribution/v3/reference FuzzParseNormalizedNamed fuzz_parsed_normalized_named
compile_go_fuzzer github.com/distribution/distribution/v3/registry/api/v2 FuzzParseForwardedHeader fuzz_parse_forwarded_header
