window.BENCHMARK_DATA = {
  "lastUpdate": 1744809929324,
  "repoUrl": "https://github.com/langgenius/dify-plugin-daemon",
  "entries": {
    "Go Benchmark": [
      {
        "commit": {
          "author": {
            "name": "langgenius",
            "username": "langgenius"
          },
          "committer": {
            "name": "langgenius",
            "username": "langgenius"
          },
          "id": "f50d9d3e9a52bde28c1111132856bbd691620172",
          "message": "feat: add benchmark workflow",
          "timestamp": "2025-04-16T11:41:26Z",
          "url": "https://github.com/langgenius/dify-plugin-daemon/pull/216/commits/f50d9d3e9a52bde28c1111132856bbd691620172"
        },
        "date": 1744808843098,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 35.49,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "33513736 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.49,
            "unit": "ns/op",
            "extra": "33513736 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "33513736 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "33513736 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode",
            "value": 27.51,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "43642790 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode - ns/op",
            "value": 27.51,
            "unit": "ns/op",
            "extra": "43642790 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "43642790 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "43642790 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode",
            "value": 26.76,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "44822308 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode - ns/op",
            "value": 26.76,
            "unit": "ns/op",
            "extra": "44822308 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "44822308 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "44822308 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode",
            "value": 12.83,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "92968137 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode - ns/op",
            "value": 12.83,
            "unit": "ns/op",
            "extra": "92968137 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "92968137 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "92968137 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode",
            "value": 36.68,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "32454444 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode - ns/op",
            "value": 36.68,
            "unit": "ns/op",
            "extra": "32454444 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "32454444 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "32454444 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode",
            "value": 514783,
            "unit": "ns/op\t  275522 B/op\t     615 allocs/op",
            "extra": "2349 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode - ns/op",
            "value": 514783,
            "unit": "ns/op",
            "extra": "2349 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode - B/op",
            "value": 275522,
            "unit": "B/op",
            "extra": "2349 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode - allocs/op",
            "value": 615,
            "unit": "allocs/op",
            "extra": "2349 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode",
            "value": 1644961,
            "unit": "ns/op\t  372340 B/op\t     707 allocs/op",
            "extra": "674 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode - ns/op",
            "value": 1644961,
            "unit": "ns/op",
            "extra": "674 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode - B/op",
            "value": 372340,
            "unit": "B/op",
            "extra": "674 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode - allocs/op",
            "value": 707,
            "unit": "allocs/op",
            "extra": "674 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode",
            "value": 258208,
            "unit": "ns/op\t   66123 B/op\t       2 allocs/op",
            "extra": "4467 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode - ns/op",
            "value": 258208,
            "unit": "ns/op",
            "extra": "4467 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode - B/op",
            "value": 66123,
            "unit": "B/op",
            "extra": "4467 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4467 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode",
            "value": 333721,
            "unit": "ns/op\t  139977 B/op\t     725 allocs/op",
            "extra": "3649 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode - ns/op",
            "value": 333721,
            "unit": "ns/op",
            "extra": "3649 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode - B/op",
            "value": 139977,
            "unit": "B/op",
            "extra": "3649 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode - allocs/op",
            "value": 725,
            "unit": "allocs/op",
            "extra": "3649 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode",
            "value": 757630,
            "unit": "ns/op\t   50627 B/op\t    2423 allocs/op",
            "extra": "1580 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode - ns/op",
            "value": 757630,
            "unit": "ns/op",
            "extra": "1580 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode - B/op",
            "value": 50627,
            "unit": "B/op",
            "extra": "1580 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode - allocs/op",
            "value": 2423,
            "unit": "allocs/op",
            "extra": "1580 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode",
            "value": 7897430,
            "unit": "ns/op\t  481096 B/op\t    8198 allocs/op",
            "extra": "152 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode - ns/op",
            "value": 7897430,
            "unit": "ns/op",
            "extra": "152 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode - B/op",
            "value": 481096,
            "unit": "B/op",
            "extra": "152 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode - allocs/op",
            "value": 8198,
            "unit": "allocs/op",
            "extra": "152 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode",
            "value": 815754,
            "unit": "ns/op\t  175880 B/op\t    3937 allocs/op",
            "extra": "1450 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode - ns/op",
            "value": 815754,
            "unit": "ns/op",
            "extra": "1450 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode - B/op",
            "value": 175880,
            "unit": "B/op",
            "extra": "1450 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode - allocs/op",
            "value": 3937,
            "unit": "allocs/op",
            "extra": "1450 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode",
            "value": 605492,
            "unit": "ns/op\t  300323 B/op\t    6357 allocs/op",
            "extra": "1976 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode - ns/op",
            "value": 605492,
            "unit": "ns/op",
            "extra": "1976 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode - B/op",
            "value": 300323,
            "unit": "B/op",
            "extra": "1976 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode - allocs/op",
            "value": 6357,
            "unit": "allocs/op",
            "extra": "1976 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode",
            "value": 1672126,
            "unit": "ns/op\t  917291 B/op\t   14219 allocs/op",
            "extra": "710 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode - ns/op",
            "value": 1672126,
            "unit": "ns/op",
            "extra": "710 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode - B/op",
            "value": 917291,
            "unit": "B/op",
            "extra": "710 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode - allocs/op",
            "value": 14219,
            "unit": "allocs/op",
            "extra": "710 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode",
            "value": 11.13,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode - ns/op",
            "value": 11.13,
            "unit": "ns/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode",
            "value": 13.89,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "85207186 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode - ns/op",
            "value": 13.89,
            "unit": "ns/op",
            "extra": "85207186 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "85207186 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "85207186 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read",
            "value": 402.6,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "2980330 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read - ns/op",
            "value": 402.6,
            "unit": "ns/op",
            "extra": "2980330 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "2980330 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "2980330 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read",
            "value": 12414,
            "unit": "ns/op\t    4040 B/op\t     252 allocs/op",
            "extra": "153316 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read - ns/op",
            "value": 12414,
            "unit": "ns/op",
            "extra": "153316 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read - B/op",
            "value": 4040,
            "unit": "B/op",
            "extra": "153316 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read - allocs/op",
            "value": 252,
            "unit": "allocs/op",
            "extra": "153316 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "45712896+Yeuoly@users.noreply.github.com",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2d216df27a465c715d4ca52ef8a6727286067be6",
          "message": "feat: add benchmark workflow (#216)\n\n* feat: add benchmark workflow\n\n- Introduced a GitHub Actions workflow for benchmarking Go code, triggered on pushes to the main branch and pull requests.\n- Added a benchmark test for the stream package to measure performance of the Write method under concurrent conditions.\n\n* fix: setup license\n\n* fix: exclude non-benchmark\n\n* fix: stash license\n\n* chore: update triggers\n\n* update README",
          "timestamp": "2025-04-16T21:23:54+08:00",
          "tree_id": "b940f60a81b8b5254fa01798a1d38b2009206013",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/2d216df27a465c715d4ca52ef8a6727286067be6"
        },
        "date": 1744809928496,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.54,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "32197672 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.54,
            "unit": "ns/op",
            "extra": "32197672 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "32197672 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "32197672 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode",
            "value": 27.52,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "44226235 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode - ns/op",
            "value": 27.52,
            "unit": "ns/op",
            "extra": "44226235 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "44226235 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Encode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "44226235 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode",
            "value": 26.79,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "38762364 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode - ns/op",
            "value": 26.79,
            "unit": "ns/op",
            "extra": "38762364 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "38762364 times\n4 procs"
          },
          {
            "name": "BenchmarkAscii85/Decode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "38762364 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode",
            "value": 12.79,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "93784486 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode - ns/op",
            "value": 12.79,
            "unit": "ns/op",
            "extra": "93784486 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "93784486 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Encode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "93784486 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode",
            "value": 36.25,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "32975262 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode - ns/op",
            "value": 36.25,
            "unit": "ns/op",
            "extra": "32975262 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "32975262 times\n4 procs"
          },
          {
            "name": "BenchmarkBase64/Decode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "32975262 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode",
            "value": 538328,
            "unit": "ns/op\t  275521 B/op\t     615 allocs/op",
            "extra": "2370 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode - ns/op",
            "value": 538328,
            "unit": "ns/op",
            "extra": "2370 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode - B/op",
            "value": 275521,
            "unit": "B/op",
            "extra": "2370 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Encode - allocs/op",
            "value": 615,
            "unit": "allocs/op",
            "extra": "2370 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode",
            "value": 1612493,
            "unit": "ns/op\t  369563 B/op\t     707 allocs/op",
            "extra": "756 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode - ns/op",
            "value": 1612493,
            "unit": "ns/op",
            "extra": "756 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode - B/op",
            "value": 369563,
            "unit": "B/op",
            "extra": "756 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Encode - allocs/op",
            "value": 707,
            "unit": "allocs/op",
            "extra": "756 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode",
            "value": 259251,
            "unit": "ns/op\t   66094 B/op\t       2 allocs/op",
            "extra": "4453 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode - ns/op",
            "value": 259251,
            "unit": "ns/op",
            "extra": "4453 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode - B/op",
            "value": 66094,
            "unit": "B/op",
            "extra": "4453 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Encode - allocs/op",
            "value": 2,
            "unit": "allocs/op",
            "extra": "4453 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode",
            "value": 335574,
            "unit": "ns/op\t  139968 B/op\t     725 allocs/op",
            "extra": "3585 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode - ns/op",
            "value": 335574,
            "unit": "ns/op",
            "extra": "3585 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode - B/op",
            "value": 139968,
            "unit": "B/op",
            "extra": "3585 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Encode - allocs/op",
            "value": 725,
            "unit": "allocs/op",
            "extra": "3585 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode",
            "value": 777707,
            "unit": "ns/op\t   50628 B/op\t    2423 allocs/op",
            "extra": "1576 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode - ns/op",
            "value": 777707,
            "unit": "ns/op",
            "extra": "1576 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode - B/op",
            "value": 50628,
            "unit": "B/op",
            "extra": "1576 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Msgpack_Decode - allocs/op",
            "value": 2423,
            "unit": "allocs/op",
            "extra": "1576 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode",
            "value": 8041374,
            "unit": "ns/op\t  481102 B/op\t    8198 allocs/op",
            "extra": "147 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode - ns/op",
            "value": 8041374,
            "unit": "ns/op",
            "extra": "147 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode - B/op",
            "value": 481102,
            "unit": "B/op",
            "extra": "147 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Json_Decode - allocs/op",
            "value": 8198,
            "unit": "allocs/op",
            "extra": "147 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode",
            "value": 823931,
            "unit": "ns/op\t  175881 B/op\t    3937 allocs/op",
            "extra": "1424 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode - ns/op",
            "value": 823931,
            "unit": "ns/op",
            "extra": "1424 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode - B/op",
            "value": 175881,
            "unit": "B/op",
            "extra": "1424 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/CBOR_Decode - allocs/op",
            "value": 3937,
            "unit": "allocs/op",
            "extra": "1424 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode",
            "value": 612808,
            "unit": "ns/op\t  300328 B/op\t    6357 allocs/op",
            "extra": "2019 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode - ns/op",
            "value": 612808,
            "unit": "ns/op",
            "extra": "2019 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode - B/op",
            "value": 300328,
            "unit": "B/op",
            "extra": "2019 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/GOB_Decode - allocs/op",
            "value": 6357,
            "unit": "allocs/op",
            "extra": "2019 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode",
            "value": 1689613,
            "unit": "ns/op\t  917233 B/op\t   14219 allocs/op",
            "extra": "699 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode - ns/op",
            "value": 1689613,
            "unit": "ns/op",
            "extra": "699 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode - B/op",
            "value": 917233,
            "unit": "B/op",
            "extra": "699 times\n4 procs"
          },
          {
            "name": "BenchmarkMsgpackVsJson/Map_Json_Decode - allocs/op",
            "value": 14219,
            "unit": "allocs/op",
            "extra": "699 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode",
            "value": 11.08,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode - ns/op",
            "value": 11.08,
            "unit": "ns/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Encode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "100000000 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode",
            "value": 13.58,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "87519602 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode - ns/op",
            "value": 13.58,
            "unit": "ns/op",
            "extra": "87519602 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "87519602 times\n4 procs"
          },
          {
            "name": "BenchmarkHex/Decode - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "87519602 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read",
            "value": 400.8,
            "unit": "ns/op\t       0 B/op\t       0 allocs/op",
            "extra": "3003940 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read - ns/op",
            "value": 400.8,
            "unit": "ns/op",
            "extra": "3003940 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read - B/op",
            "value": 0,
            "unit": "B/op",
            "extra": "3003940 times\n4 procs"
          },
          {
            "name": "BenchmarkStdioBandWidth/Read - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "3003940 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read",
            "value": 12142,
            "unit": "ns/op\t    3921 B/op\t     244 allocs/op",
            "extra": "127126 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read - ns/op",
            "value": 12142,
            "unit": "ns/op",
            "extra": "127126 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read - B/op",
            "value": 3921,
            "unit": "B/op",
            "extra": "127126 times\n4 procs"
          },
          {
            "name": "BenchmarkStreamResponse/Read - allocs/op",
            "value": 244,
            "unit": "allocs/op",
            "extra": "127126 times\n4 procs"
          }
        ]
      }
    ]
  }
}