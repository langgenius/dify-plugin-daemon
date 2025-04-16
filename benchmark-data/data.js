window.BENCHMARK_DATA = {
  "lastUpdate": 1744808843415,
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
      }
    ]
  }
}