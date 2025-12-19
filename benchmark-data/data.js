window.BENCHMARK_DATA = {
  "lastUpdate": 1766129531188,
  "repoUrl": "https://github.com/langgenius/dify-plugin-daemon",
  "entries": {
    "Go Benchmark": [
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
          "id": "7a92f1d9ddc8b3d3dcd2c902b6bc045ae1a6373d",
          "message": "chore: remove useless benchmarks (#217)\n\n* chore: remove useless benchmarks\n\n* fix: remove tests",
          "timestamp": "2025-04-16T21:30:35+08:00",
          "tree_id": "5de622fc8f231d1406b2a2fc87ada640285650ec",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7a92f1d9ddc8b3d3dcd2c902b6bc045ae1a6373d"
        },
        "date": 1744810310104,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.71,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "29869750 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.71,
            "unit": "ns/op",
            "extra": "29869750 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "29869750 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "29869750 times\n4 procs"
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
          "id": "73a4d93a627ad2617af1e0d5a27330fc2a9638b9",
          "message": "benchmark/local-runtime (#219)\n\n* benchmark/local-runtime\n\n* test\n\n* fix\n\n* fix: add uv\n\n* fix: uv path\n\n* fix: get uv\n\n* done",
          "timestamp": "2025-04-17T00:03:28+08:00",
          "tree_id": "e01d4a2181c6e575829bd1a3e1b2eda8257df697",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/73a4d93a627ad2617af1e0d5a27330fc2a9638b9"
        },
        "date": 1744819612212,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.1,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "992299392 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.1,
            "unit": "ns/op",
            "extra": "992299392 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "992299392 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "992299392 times\n4 procs"
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
          "id": "ecfe13a5453e6b1bfd123cccf971b36180ffe767",
          "message": "Enhance benchmark test: disable logging for local OpenAI LLM invocation (#220)",
          "timestamp": "2025-04-17T00:11:51+08:00",
          "tree_id": "b1e483e40a3b7999ef63ebc75decfbe357c74b01",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ecfe13a5453e6b1bfd123cccf971b36180ffe767"
        },
        "date": 1744820118176,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142330820,
            "unit": "ns/op\t 1593803 B/op\t   28317 allocs/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142330820,
            "unit": "ns/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593803,
            "unit": "B/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.99,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "957398096 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.99,
            "unit": "ns/op",
            "extra": "957398096 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "957398096 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "957398096 times\n4 procs"
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
          "id": "c5cda9cbbf892894a1f87455ccf00b1d98b5500e",
          "message": "enhance(local runtime): patches prompt messages and replace json loads with pydantic (#222)\n\n- Introduced two new patch files: `0.1.1.llm.py.patch` and `0.1.1.request_reader.py.patch` to enhance the plugin's functionality.\n- Updated the `environment_python.go` file to reference the new patches and apply them conditionally based on the plugin SDK version.\n- Improved the handling of LLM usage and request reading in the plugin environment.",
          "timestamp": "2025-04-17T16:07:54+08:00",
          "tree_id": "e3d777c2d7d6f80eedeb0a811a337aed00f8546a",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/c5cda9cbbf892894a1f87455ccf00b1d98b5500e"
        },
        "date": 1744877477702,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139370703,
            "unit": "ns/op\t 1593381 B/op\t   28316 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139370703,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593381,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.08,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "987237850 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.08,
            "unit": "ns/op",
            "extra": "987237850 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "987237850 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "987237850 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liang.bowen.123@qq.com",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ae806d03f6936f45db76bf6d0165f7cbfd4877df",
          "message": "doc: update cli tool installation guidance in README.md (#227)",
          "timestamp": "2025-04-20T17:40:45+08:00",
          "tree_id": "380e119ecab8bdb178adfb9b7dfa1b92047654d7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ae806d03f6936f45db76bf6d0165f7cbfd4877df"
        },
        "date": 1745142250116,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138853888,
            "unit": "ns/op\t 1593435 B/op\t   28316 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138853888,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593435,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.3,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "972339002 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.3,
            "unit": "ns/op",
            "extra": "972339002 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "972339002 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "972339002 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liang.bowen.123@qq.com",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ef676bee4737bbb88e12347f477047da531b56e4",
          "message": "chore: align go version in GHA to 1.23 (#225)",
          "timestamp": "2025-04-20T17:43:05+08:00",
          "tree_id": "7d6b8704df63175df8065d05c39861f07eb6b559",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ef676bee4737bbb88e12347f477047da531b56e4"
        },
        "date": 1745142391344,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141527092,
            "unit": "ns/op\t 1593484 B/op\t   28316 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141527092,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593484,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.16,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "959219251 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.16,
            "unit": "ns/op",
            "extra": "959219251 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "959219251 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "959219251 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liang.bowen.123@qq.com",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d918e0292dcb8d12559e9d141f6edf326f8af76b",
          "message": "chore: modernize GHA versions (#226)",
          "timestamp": "2025-04-20T17:43:38+08:00",
          "tree_id": "87fc15a89d2f6f866577792dad054a7701d51b62",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d918e0292dcb8d12559e9d141f6edf326f8af76b"
        },
        "date": 1745142415077,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138716091,
            "unit": "ns/op\t 1593762 B/op\t   28317 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138716091,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593762,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.22,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "974120535 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.22,
            "unit": "ns/op",
            "extra": "974120535 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "974120535 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "974120535 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "distinct": true,
          "id": "102e54258be479f630d0173ef5c259bd45a280be",
          "message": "refactor",
          "timestamp": "2025-04-21T23:02:45+08:00",
          "tree_id": "b6c77fa92afb8c1d9f9108ffd58819f0193e9def",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/102e54258be479f630d0173ef5c259bd45a280be"
        },
        "date": 1745248312888,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.25,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "994294190 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.25,
            "unit": "ns/op",
            "extra": "994294190 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "994294190 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "994294190 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "distinct": true,
          "id": "d811426f64b8396575543a7cef1aa9bcc0ec35e9",
          "message": "feat: auto scale",
          "timestamp": "2025-04-21T22:56:05+08:00",
          "tree_id": "4a8f35994d815583c0f7c27a7b5058aacf6c417b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d811426f64b8396575543a7cef1aa9bcc0ec35e9"
        },
        "date": 1745248566409,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.26,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982894938 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.26,
            "unit": "ns/op",
            "extra": "982894938 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982894938 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982894938 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "45784494+lcandy2@users.noreply.github.com",
            "name": "cirtron",
            "username": "lcandy2"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e2a11dd6001b0f2862201cb9296560bc0db7c596",
          "message": "chore: update macos gitignore (#228)",
          "timestamp": "2025-04-22T13:22:00+08:00",
          "tree_id": "83108ce5cc71c27157afb90db6558f5b8d047813",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/e2a11dd6001b0f2862201cb9296560bc0db7c596"
        },
        "date": 1745299477122,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142969262,
            "unit": "ns/op\t 1593269 B/op\t   28316 allocs/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142969262,
            "unit": "ns/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593269,
            "unit": "B/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "244 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.35,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "987819278 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.35,
            "unit": "ns/op",
            "extra": "987819278 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "987819278 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "987819278 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "distinct": true,
          "id": "63279c70c1a2e02926c1f6fe48900b65d46a7f50",
          "message": "refactor",
          "timestamp": "2025-04-22T14:21:30+08:00",
          "tree_id": "428b46e19a339e83af8a9b18ad707cfc0c1ef11c",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/63279c70c1a2e02926c1f6fe48900b65d46a7f50"
        },
        "date": 1745303402172,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.55,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "976079280 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.55,
            "unit": "ns/op",
            "extra": "976079280 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "976079280 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "976079280 times\n4 procs"
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
          "id": "399c027756d60180f952f960636d8c865a95d239",
          "message": "chore: add checkout step in publish-cli workflow (#229) (#232)",
          "timestamp": "2025-04-22T16:29:27+08:00",
          "tree_id": "e0bec5020ce82f55899ae2621482ff7479095a27",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/399c027756d60180f952f960636d8c865a95d239"
        },
        "date": 1745310730915,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141091598,
            "unit": "ns/op\t 1593341 B/op\t   28316 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141091598,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593341,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.41,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "975444310 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.41,
            "unit": "ns/op",
            "extra": "975444310 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "975444310 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "975444310 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49699333+dependabot[bot]@users.noreply.github.com",
            "name": "dependabot[bot]",
            "username": "dependabot[bot]"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "300d362d1f80dc550b5d92d8bcde2d756f9c26f8",
          "message": "chore(deps): bump golang.org/x/net from 0.34.0 to 0.38.0 (#233)\n\nBumps [golang.org/x/net](https://github.com/golang/net) from 0.34.0 to 0.38.0.\n- [Commits](https://github.com/golang/net/compare/v0.34.0...v0.38.0)\n\n---\nupdated-dependencies:\n- dependency-name: golang.org/x/net\n  dependency-version: 0.38.0\n  dependency-type: indirect\n...\n\nSigned-off-by: dependabot[bot] <support@github.com>\nCo-authored-by: dependabot[bot] <49699333+dependabot[bot]@users.noreply.github.com>",
          "timestamp": "2025-04-22T16:45:58+08:00",
          "tree_id": "78cfd1e7f060bd1897672b9b5bb1d92b5709ab04",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/300d362d1f80dc550b5d92d8bcde2d756f9c26f8"
        },
        "date": 1745311760414,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143109887,
            "unit": "ns/op\t 1593653 B/op\t   28317 allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143109887,
            "unit": "ns/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593653,
            "unit": "B/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.12,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "953646408 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.12,
            "unit": "ns/op",
            "extra": "953646408 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "953646408 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "953646408 times\n4 procs"
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
          "id": "8ff2949924fdffc9082b8520c00defa921db34f7",
          "message": "chore: update build-push workflow to skip builds on pull requests (#233) (#234)\n\n* chore: update build-push workflow to skip builds on pull requests (#233)\n\n- Modified the conditional for the build job to skip execution on pull requests when the `skip_on_pr` flag is set to true for specific services.\n\n* optimize",
          "timestamp": "2025-04-22T16:55:07+08:00",
          "tree_id": "17aeeb98d005e82c01b88e43006a88d251ecbabc",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/8ff2949924fdffc9082b8520c00defa921db34f7"
        },
        "date": 1745312269355,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143251771,
            "unit": "ns/op\t 1593636 B/op\t   28317 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143251771,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593636,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.38,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982955383 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.38,
            "unit": "ns/op",
            "extra": "982955383 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982955383 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982955383 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49699333+dependabot[bot]@users.noreply.github.com",
            "name": "dependabot[bot]",
            "username": "dependabot[bot]"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9da6137926c750b9bd9382c1c7ad1420f931a2f5",
          "message": "chore(deps): bump github.com/redis/go-redis/v9 from 9.5.3 to 9.5.5 (#235)\n\nBumps [github.com/redis/go-redis/v9](https://github.com/redis/go-redis) from 9.5.3 to 9.5.5.\n- [Release notes](https://github.com/redis/go-redis/releases)\n- [Changelog](https://github.com/redis/go-redis/blob/master/CHANGELOG.md)\n- [Commits](https://github.com/redis/go-redis/compare/v9.5.3...v9.5.5)\n\n---\nupdated-dependencies:\n- dependency-name: github.com/redis/go-redis/v9\n  dependency-version: 9.5.5\n  dependency-type: direct:production\n...\n\nSigned-off-by: dependabot[bot] <support@github.com>\nCo-authored-by: dependabot[bot] <49699333+dependabot[bot]@users.noreply.github.com>",
          "timestamp": "2025-04-22T17:31:20+08:00",
          "tree_id": "66ebda4fe88b1a1698498425f02b9d226cf46f41",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9da6137926c750b9bd9382c1c7ad1420f931a2f5"
        },
        "date": 1745314485781,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141087930,
            "unit": "ns/op\t 1593299 B/op\t   28316 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141087930,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593299,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.87,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "989154942 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.87,
            "unit": "ns/op",
            "extra": "989154942 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "989154942 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "989154942 times\n4 procs"
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
          "id": "c0498dda3a4b43549bb6ed51537c05f1aca807d0",
          "message": "Merge branch 'main' into feat/auto-scale",
          "timestamp": "2025-04-22T19:41:21+08:00",
          "tree_id": "d2e247871432995069d4e73310368a922fa381d9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/c0498dda3a4b43549bb6ed51537c05f1aca807d0"
        },
        "date": 1745322727577,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.06,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "990543537 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.06,
            "unit": "ns/op",
            "extra": "990543537 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "990543537 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "990543537 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "distinct": true,
          "id": "f9bedbad4eccbef13bc46bae72e7f28cf7e1db92",
          "message": "fix: ci",
          "timestamp": "2025-04-22T19:58:54+08:00",
          "tree_id": "ee0ae8de7713672068125fcf748f84456793ee2c",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/f9bedbad4eccbef13bc46bae72e7f28cf7e1db92"
        },
        "date": 1745323861737,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.83,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "990959812 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.83,
            "unit": "ns/op",
            "extra": "990959812 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "990959812 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "990959812 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "distinct": true,
          "id": "01ef5f27a311e52cec26e4cabb3d83c56e72303a",
          "message": "fix: ci",
          "timestamp": "2025-04-22T20:04:14+08:00",
          "tree_id": "0b17b255126018b9035bf2c2a37851011acf17d9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/01ef5f27a311e52cec26e4cabb3d83c56e72303a"
        },
        "date": 1745324235614,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 37.16,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "971695618 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.16,
            "unit": "ns/op",
            "extra": "971695618 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "971695618 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "971695618 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "committer": {
            "email": "admin@srmxy.cn",
            "name": "Yeuoly",
            "username": "Yeuoly"
          },
          "distinct": true,
          "id": "225742aa94a738c7c2ef41f11b843ec00b984480",
          "message": "optimize: load balancing",
          "timestamp": "2025-04-22T20:40:48+08:00",
          "tree_id": "1702bb1abb3624acad69d8cbf1ffd6422bf49d2c",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/225742aa94a738c7c2ef41f11b843ec00b984480"
        },
        "date": 1745326227813,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.15,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "989561757 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.15,
            "unit": "ns/op",
            "extra": "989561757 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "989561757 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "989561757 times\n4 procs"
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
          "id": "ea67e810ddcd795466e4cd1a8fcedda19454a033",
          "message": "feat: add backward compatibility for LLMResultChunk and new test case (#240)\n\n- Implemented a compatibility layer in LLMResultChunk to ensure backward compatibility with the old format by adding a `PromptMessages` field.\n- Added a new test case `TestLLMResultChunkCompatibility` to verify the JSON marshaling behavior of the updated LLMResultChunk structure.",
          "timestamp": "2025-04-24T13:37:46+08:00",
          "tree_id": "0724b935e2d4193c30d686f5ec5b1660b13e9d1b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ea67e810ddcd795466e4cd1a8fcedda19454a033"
        },
        "date": 1745473231934,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141768634,
            "unit": "ns/op\t 1593368 B/op\t   28316 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141768634,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593368,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.97,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "989799157 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.97,
            "unit": "ns/op",
            "extra": "989799157 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "989799157 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "989799157 times\n4 procs"
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
          "id": "9730b7dc63bd9c1eb05cf2c0f36755738a162868",
          "message": "tests: enhance integration testing for specific plugins (#242)\n\n* feat: enhance integration testing for specific plugins\n\n- Added integration tests for the official agent, including test data for agent strategy invocation.\n- Introduced JSON schema generation and validation utilities to ensure proper request formatting.\n- Enhanced mock invocation handling in the plugin manager to support tool parameters.\n- Added new test utilities for simulating OpenAI server responses and managing plugin runtime.\n\n* fix: update RunOnce function to return response stream and enhance integration test\n\n- Modified the RunOnce function to return a response stream instead of an error, allowing for better handling of streamed responses.\n- Updated the integration test for the official agent to read from the response stream, ensuring proper validation of the agent strategy invocation.",
          "timestamp": "2025-04-24T16:33:50+08:00",
          "tree_id": "3fe97f790194dc28f74a7ba43f573fc45c1bd575",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9730b7dc63bd9c1eb05cf2c0f36755738a162868"
        },
        "date": 1745483793372,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141508021,
            "unit": "ns/op\t 1594183 B/op\t   28317 allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141508021,
            "unit": "ns/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1594183,
            "unit": "B/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.65,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "987052456 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.65,
            "unit": "ns/op",
            "extra": "987052456 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "987052456 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "987052456 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "mr.nikoru918@gmail.com",
            "name": "Hironori Yamamoto",
            "username": "hiro-o918"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2406e0a09ff0826b817ed520a271c1ab3805c61a",
          "message": "feat: implement gcs storage (#237)\n\nCo-authored-by: Hironori Yamamoto <hironori-yamamoto@m3.com>",
          "timestamp": "2025-04-25T13:29:08+08:00",
          "tree_id": "f8adb69f0a1b41e41f87595b62890123cf0d56ba",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/2406e0a09ff0826b817ed520a271c1ab3805c61a"
        },
        "date": 1745559184486,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140664000,
            "unit": "ns/op\t 1593377 B/op\t   28316 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140664000,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593377,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.92,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "985532454 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.92,
            "unit": "ns/op",
            "extra": "985532454 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "985532454 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "985532454 times\n4 procs"
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
          "id": "00730b67c3841a32ac23c418b0f9274fabd64e9e",
          "message": "refactor: replace baseSSEService with baseSSEWithSession in multiple service files (#244)\n\n- Introduced baseSSEWithSession to streamline session management across various service functions.\n- Updated InvokeAgentStrategy, InvokeLLM, InvokeTextEmbedding, InvokeRerank, InvokeTTS, InvokeSpeech2Text, InvokeModeration, ValidateProviderCredentials, ValidateModelCredentials, GetTTSModelVoices, GetTextEmbeddingNumTokens, GetAIModelSchema, and GetLLMNumTokens to utilize the new session handling approach, improving code readability and maintainability.",
          "timestamp": "2025-04-25T18:34:46+08:00",
          "tree_id": "af963447ad594856d965f8e553f31a2663685585",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/00730b67c3841a32ac23c418b0f9274fabd64e9e"
        },
        "date": 1745577446067,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139231238,
            "unit": "ns/op\t 1593705 B/op\t   28316 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139231238,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593705,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.12,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "983818504 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.12,
            "unit": "ns/op",
            "extra": "983818504 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "983818504 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "983818504 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "taeyoung613@gmail.com",
            "name": "Taeyoung Park",
            "username": "taeyoungpark"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5d83258371b0a051d50ec4fbbdfa99a32dffcf4e",
          "message": " Add optional Redis username authentication (#218)\n\n* You can include username for Redis Auth\n\n* Redis auth with username - test code update\n\n* fix: tests\n\n---------\n\nCo-authored-by: Yeuoly <admin@srmxy.cn>",
          "timestamp": "2025-04-25T20:42:14+08:00",
          "tree_id": "716ade3701822df44166329021b78c2aed1bc307",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/5d83258371b0a051d50ec4fbbdfa99a32dffcf4e"
        },
        "date": 1745585110259,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 148297152,
            "unit": "ns/op\t 1593570 B/op\t   28317 allocs/op",
            "extra": "232 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 148297152,
            "unit": "ns/op",
            "extra": "232 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593570,
            "unit": "B/op",
            "extra": "232 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "232 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.57,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "946875913 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.57,
            "unit": "ns/op",
            "extra": "946875913 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "946875913 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "946875913 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "99897242+chiehw@users.noreply.github.com",
            "name": "Chieh Wang",
            "username": "chiehw"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1121069521aeb714184b52957ac68a94bf2cf2cf",
          "message": "feat: Add Citations and Attributions to Agent Node (#236)\n\n* feat: support var-selector params & surpport retriever_resources messsage\n\n* opt: rename var-selector to any",
          "timestamp": "2025-04-27T13:41:21+08:00",
          "tree_id": "ec9b2199b254653463337293e475fda490a75dbb",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1121069521aeb714184b52957ac68a94bf2cf2cf"
        },
        "date": 1745732646855,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137816806,
            "unit": "ns/op\t 1593256 B/op\t   28316 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137816806,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593256,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.6,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "994132584 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.6,
            "unit": "ns/op",
            "extra": "994132584 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "994132584 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "994132584 times\n4 procs"
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
          "id": "d80630acd59e02f54812aed197853cfa0034dbea",
          "message": "feat: implement OAuth functionality (#245)\n\n* feat: implement OAuth functionality in plugin daemon\n\n- Added OAuth service methods for getting authorization URLs and credentials.\n- Updated access types to include OAuth-related actions.\n- Created new controller for handling OAuth requests.\n- Introduced entities for OAuth results and requests.\n- Enhanced plugin entities to support OAuth schema in tool declarations.\n\n* feat: add OAuth endpoints for authorization and credentials retrieval\n\n- Introduced new POST endpoints for obtaining authorization URLs and credentials in the OAuth controller.\n- Enhanced the plugin dispatch group to include these new OAuth routes, improving integration with OAuth services.",
          "timestamp": "2025-04-27T13:59:57+08:00",
          "tree_id": "f72bb5e53f9732e5f93cd2568f99daa173ee0a29",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d80630acd59e02f54812aed197853cfa0034dbea"
        },
        "date": 1745733764499,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138558126,
            "unit": "ns/op\t 1593650 B/op\t   28316 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138558126,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593650,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.33,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997990780 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.33,
            "unit": "ns/op",
            "extra": "997990780 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997990780 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997990780 times\n4 procs"
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
          "id": "d13fb55f4f0b387432cd9b9f6567843e2431cb41",
          "message": "Refactor/oauth parameters (#247)\n\n* feat: implement OAuth functionality in plugin daemon\n\n- Added OAuth service methods for getting authorization URLs and credentials.\n- Updated access types to include OAuth-related actions.\n- Created new controller for handling OAuth requests.\n- Introduced entities for OAuth results and requests.\n- Enhanced plugin entities to support OAuth schema in tool declarations.\n\n* feat: add OAuth endpoints for authorization and credentials retrieval\n\n- Introduced new POST endpoints for obtaining authorization URLs and credentials in the OAuth controller.\n- Enhanced the plugin dispatch group to include these new OAuth routes, improving integration with OAuth services.\n\n* refactor: clean up OAuth entity and request structures\n\n- Removed unused validation logic and constants from the OAuth entity file.\n- Added required provider field to OAuth request structures for better validation and clarity.",
          "timestamp": "2025-04-27T14:35:43+08:00",
          "tree_id": "1a7579c9eac2b3db6d3770632509a6e7331a6161",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d13fb55f4f0b387432cd9b9f6567843e2431cb41"
        },
        "date": 1745735917336,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138948407,
            "unit": "ns/op\t 1593024 B/op\t   28316 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138948407,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593024,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.23,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982843048 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.23,
            "unit": "ns/op",
            "extra": "982843048 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982843048 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982843048 times\n4 procs"
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
          "id": "ee540bc2209bca6792b0f8d62b8f37ce27d16ae3",
          "message": "feat: enhance plugin initialization with configurable parameters (#248)\n\n* feat: enhance plugin initialization with configurable parameters\n\n- Added new flags for plugin initialization, allowing users to specify author, name, description, and various permissions.\n- Implemented InitPluginWithFlags function to handle the new parameters and validate input.\n- Introduced methods to set category, language, and minimal Dify version within the plugin model.\n- Enhanced profile management by adding methods to set author and name directly.\n- Improved category and language selection with dedicated setter methods.\n\n* add enum\n\n* add tests",
          "timestamp": "2025-04-27T17:29:26+08:00",
          "tree_id": "7def44d7dd36c3dd75523a1c7afee2d226d350e3",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ee540bc2209bca6792b0f8d62b8f37ce27d16ae3"
        },
        "date": 1745746333715,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139648779,
            "unit": "ns/op\t 1593845 B/op\t   28316 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139648779,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593845,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.21,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982338816 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.21,
            "unit": "ns/op",
            "extra": "982338816 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982338816 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982338816 times\n4 procs"
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
          "id": "ad71e7ae261709bbdbfb7e07271a1af8ab5229d6",
          "message": "refactor: move validation logic for plugin initialization under quick flag (#249)\n\n- Updated the InitPluginWithFlags function to conditionally validate the plugin name, author, and description only when the quick flag is set.\n- This change improves the flexibility of the plugin initialization process by allowing bypassing validation in certain scenarios.",
          "timestamp": "2025-04-27T17:59:03+08:00",
          "tree_id": "6c8c32a6ae453de34c6479d801cfe6e2660ff18f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ad71e7ae261709bbdbfb7e07271a1af8ab5229d6"
        },
        "date": 1745748112678,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138136288,
            "unit": "ns/op\t 1593193 B/op\t   28316 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138136288,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593193,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.82,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997958934 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.82,
            "unit": "ns/op",
            "extra": "997958934 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997958934 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997958934 times\n4 procs"
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
          "id": "18a251a82313a00c358b8ee6fcff9054b21eef11",
          "message": "fix: update moderation model docstring to reflect plugin name correctly (#250)\n\n- Modified the docstring in the moderation model to use the correct format for the plugin name, enhancing clarity and consistency in the documentation.",
          "timestamp": "2025-04-27T18:11:37+08:00",
          "tree_id": "28c7e7f0e965508ec573608f81879bd7c09b8f41",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/18a251a82313a00c358b8ee6fcff9054b21eef11"
        },
        "date": 1745748862132,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138327416,
            "unit": "ns/op\t 1593543 B/op\t   28317 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138327416,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593543,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.18,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982733634 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.18,
            "unit": "ns/op",
            "extra": "982733634 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982733634 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982733634 times\n4 procs"
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
          "id": "61bd7d9c8a13dcb5320ad5f6db7190e7a646c98c",
          "message": "feat: add configurable stdio buffer sizes for plugins (#256)\n\n- Introduced new configuration options for plugin stdio buffer sizes in the .env.example file.\n- Updated the PluginManager and LocalPluginRuntime to utilize these new buffer size settings.\n- Enhanced the stdioHolder to accept buffer size configurations, improving plugin output handling.\n- Modified related tests to accommodate the new stdioHolder configuration structure.",
          "timestamp": "2025-04-30T15:28:34+08:00",
          "tree_id": "f9a043e96f5e0a3338372cf7e33acddc96e3177f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/61bd7d9c8a13dcb5320ad5f6db7190e7a646c98c"
        },
        "date": 1745998347479,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138413473,
            "unit": "ns/op\t 1593460 B/op\t   28316 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138413473,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593460,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.17,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "993655850 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.17,
            "unit": "ns/op",
            "extra": "993655850 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "993655850 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "993655850 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "quicksandzn@gmail.com",
            "name": "quicksand",
            "username": "quicksandznzn"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2858ac210cd1573dd11e700aa81bd3a4edb39c7b",
          "message": "optimize (db): add db pool config (#251)",
          "timestamp": "2025-04-30T15:36:35+08:00",
          "tree_id": "6a224c01f5ca441033f09daa3e17eda81d05ec32",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/2858ac210cd1573dd11e700aa81bd3a4edb39c7b"
        },
        "date": 1745998825855,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142518160,
            "unit": "ns/op\t 1593524 B/op\t   28317 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142518160,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593524,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.22,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "928951952 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.22,
            "unit": "ns/op",
            "extra": "928951952 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "928951952 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "928951952 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "slm_1990@126.com",
            "name": "Good Wood",
            "username": "LeeeeeeM"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "977665e73b600c247b132bc958284b7f296c5231",
          "message": "feat: add read & write timeout config (#259)\n\n* feat: add read & write timeout config\n\n* refactor: update Dify invocation configuration to use structured payload\n\n- Changed the Dify invocation daemon to accept a structured payload for initialization, improving clarity and maintainability.\n- Updated related configuration variables in the .env.example file to reflect the new naming convention for backwards invocation timeouts.\n- Adjusted tests and plugin manager to accommodate the new payload structure.\n\n---------\n\nCo-authored-by: Yeuoly <admin@srmxy.cn>",
          "timestamp": "2025-05-06T13:08:02+08:00",
          "tree_id": "701c71fc4474379123d20a2b32a97e901f886b1d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/977665e73b600c247b132bc958284b7f296c5231"
        },
        "date": 1746508307980,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139633054,
            "unit": "ns/op\t 1593429 B/op\t   28316 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139633054,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593429,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.91,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981067863 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.91,
            "unit": "ns/op",
            "extra": "981067863 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981067863 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981067863 times\n4 procs"
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
          "id": "1ebe7fdf655939ccab8eda12c1892919b5af342f",
          "message": "fix: improve line processing in StdioRequestReader (#260)\n\n* fix: improve line processing in StdioRequestReader\n\n- Replaced synchronous stdin reading with asynchronous data retrieval.\n- Enhanced line processing to skip empty lines and handle incomplete data more effectively, improving overall robustness of the request reading mechanism.\n\n* feat: add asynchronous reading to StdioRequestReader\n\n- Implemented the _read_async method to read data from stdin in 64KB chunks, optimizing data retrieval.\n- This enhancement improves the efficiency of the request reading process in the StdioRequestReader class.",
          "timestamp": "2025-05-06T15:09:54+08:00",
          "tree_id": "7e3e575a172e0eb95f1e143b6fa4012495c4bdea",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1ebe7fdf655939ccab8eda12c1892919b5af342f"
        },
        "date": 1746515632303,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141888694,
            "unit": "ns/op\t 1593458 B/op\t   28316 allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141888694,
            "unit": "ns/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593458,
            "unit": "B/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.09,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997479056 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.09,
            "unit": "ns/op",
            "extra": "997479056 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997479056 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997479056 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "62681807+bravomark@users.noreply.github.com",
            "name": "bravomark",
            "username": "bravomark"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7492a3d8cdd1f028199de67e6671ef8a4c2c559a",
          "message": "feat: Support Alibaba Cloud OSS (#261)\n\n* feat: support aliyun OSS\n\n* feat: support aliyun OSS",
          "timestamp": "2025-05-07T16:46:18+08:00",
          "tree_id": "0523fa3668cd4721049520dbeed9761767ebdcc0",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7492a3d8cdd1f028199de67e6671ef8a4c2c559a"
        },
        "date": 1746607811207,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138440989,
            "unit": "ns/op\t 1593928 B/op\t   28317 allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138440989,
            "unit": "ns/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593928,
            "unit": "B/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.88,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "996924553 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.88,
            "unit": "ns/op",
            "extra": "996924553 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "996924553 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "996924553 times\n4 procs"
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
          "id": "d02885674a3e391e2941b838403d09ba32055b84",
          "message": "feat: add repository URL support for plugin manifest (#262)\n\n* feat: add repository URL support for plugin manifest\n\n- Introduced a new optional flag for specifying the plugin repository URL during initialization.\n- Updated the InitPluginWithFlags function to handle the new repository parameter.\n- Enhanced profile management to include repository input.\n- Modified related tests to validate the new repository functionality.\n\n* fix: improve input validation in profile checkRule method\n\n- Updated the checkRule method to ensure cursor is within valid range before checking for empty input values. This change prevents potential out-of-bounds errors and enhances the robustness of the input validation process.",
          "timestamp": "2025-05-07T18:35:38+08:00",
          "tree_id": "a495b8a0e4baa16a77a05a3f1a52e4f0601b4d4b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d02885674a3e391e2941b838403d09ba32055b84"
        },
        "date": 1746614306639,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141632802,
            "unit": "ns/op\t 1593269 B/op\t   28316 allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141632802,
            "unit": "ns/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593269,
            "unit": "B/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.19,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "976410202 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.19,
            "unit": "ns/op",
            "extra": "976410202 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "976410202 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "976410202 times\n4 procs"
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
          "id": "7a7e23fcb13955562ab9c8085a021090960beec7",
          "message": "fix: update stdio buffer size configuration in stdioHolder (#273)\n\n- Adjusted the buffer size settings in the stdioHolder to utilize configurable values for stdout buffer size and maximum buffer size, enhancing flexibility in plugin output handling.",
          "timestamp": "2025-05-09T12:28:14+08:00",
          "tree_id": "1430c0b38c3c6b2df8a2f10cbdce98196668e9ac",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7a7e23fcb13955562ab9c8085a021090960beec7"
        },
        "date": 1746765063076,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139993821,
            "unit": "ns/op\t 1593731 B/op\t   28317 allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139993821,
            "unit": "ns/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593731,
            "unit": "B/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.05,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.05,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "f99e0e8119720eb94972772cc73c6c3ab23236ee",
          "message": "feat: add input validation for installed bucket path (#264)\n\n- Implemented validation in NewInstalledBucket to log warnings for non-alphanumeric starting characters and empty paths for installedPath.\n- Updated parameter naming for consistency across methods in InstalledBucket, enhancing code clarity.",
          "timestamp": "2025-05-09T12:28:47+08:00",
          "tree_id": "a0b65f858f6e9a6493cbf440b23c1bdc29178c1e",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/f99e0e8119720eb94972772cc73c6c3ab23236ee"
        },
        "date": 1746765092334,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141482491,
            "unit": "ns/op\t 1593834 B/op\t   28316 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141482491,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593834,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.16,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "979537180 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.16,
            "unit": "ns/op",
            "extra": "979537180 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "979537180 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "979537180 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liangbowen@gf.com.cn",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "de29e4ba1870a79979399cb55087e431ef09a2fa",
          "message": "update (#271)",
          "timestamp": "2025-05-09T12:38:47+08:00",
          "tree_id": "ee95afa182e48a71806ceace1ae1524d781c6360",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/de29e4ba1870a79979399cb55087e431ef09a2fa"
        },
        "date": 1746765691152,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139092967,
            "unit": "ns/op\t 1593440 B/op\t   28316 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139092967,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593440,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.31,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981873210 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.31,
            "unit": "ns/op",
            "extra": "981873210 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981873210 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981873210 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liangbowen@gf.com.cn",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "07e92a85b5f4ea3213099911e349a1a3e60d0b94",
          "message": "add scripts (#272)",
          "timestamp": "2025-05-09T12:40:04+08:00",
          "tree_id": "0138f70c76203ab0dacdf0d9ec3c854aa64a334a",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/07e92a85b5f4ea3213099911e349a1a3e60d0b94"
        },
        "date": 1746765768819,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140803568,
            "unit": "ns/op\t 1593368 B/op\t   28316 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140803568,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593368,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.97,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "995853187 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.97,
            "unit": "ns/op",
            "extra": "995853187 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "995853187 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "995853187 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liangbowen@gf.com.cn",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5193a3ff4d601d8197c9523adea5d1cf42ad7830",
          "message": "update (#266)",
          "timestamp": "2025-05-09T12:41:39+08:00",
          "tree_id": "7ce847af7116ac1a3d780721df68f96e7dee9ee9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/5193a3ff4d601d8197c9523adea5d1cf42ad7830"
        },
        "date": 1746765864623,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142088926,
            "unit": "ns/op\t 1593818 B/op\t   28317 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142088926,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593818,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.74,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984273422 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.74,
            "unit": "ns/op",
            "extra": "984273422 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984273422 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984273422 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liangbowen@gf.com.cn",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e655881e6fa6c15d72f649449b7798e0b3551b39",
          "message": "feat: support REMOTE_INSTALL_URL in plugin cli template (#279)\n\n* set REMOTE_INSTALL_URL in .env.example\n\n* update GUIDE.md\n\n* remove REMOTE_INSTALL_PORT config in GUIDE.md\n\n* remove REMOTE_INSTALL_PORT config in GUIDE.md\n\n* update REMOTE_INSTALL_URL config in GUIDE.md\n\n* update REMOTE_INSTALL_URL config in GUIDE.md",
          "timestamp": "2025-05-13T14:24:23+08:00",
          "tree_id": "cbdc9d1ff063e36ee86dfa8b2e0f021a9d095855",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/e655881e6fa6c15d72f649449b7798e0b3551b39"
        },
        "date": 1747117698154,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139607622,
            "unit": "ns/op\t 1593594 B/op\t   28317 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139607622,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593594,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.22,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "977879695 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.22,
            "unit": "ns/op",
            "extra": "977879695 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "977879695 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "977879695 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "liangbowen@gf.com.cn",
            "name": "Bowen Liang",
            "username": "bowenliang123"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "fa2ac6dd2cd82df079743106dc357486379785a2",
          "message": "dep: bump dify_plugin in plugin dependency template to 0.2.x (#278)\n\n* bump dify_plugin to 0.2.x\n\n* dify_plugin>=0.2.0,<0.3.0",
          "timestamp": "2025-05-13T17:13:56+08:00",
          "tree_id": "fba0b2a9e2e0d2d9b3ee218b8e80497d6c58c4d5",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/fa2ac6dd2cd82df079743106dc357486379785a2"
        },
        "date": 1747127868669,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140682954,
            "unit": "ns/op\t 1593065 B/op\t   28316 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140682954,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593065,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.74,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.74,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "d884cab8aec86ab9a27c5c3ca1ebd5686c8e3db7",
          "message": "feat: add plugin reinstallation functionality and admin API support (#285)\n\n* feat: add plugin reinstallation functionality and admin API support\n\n- Implemented ReinstallToAWSFromPkg method to allow reinstallation of plugins on AWS Lambda, updating function URL and name.\n- Added clearServerlessRuntimeCache method to manage serverless runtime cache.\n- Enhanced LaunchPlugin to support an ignoreIdempotent flag for forced reinstallation.\n- Introduced admin API endpoints for plugin reinstallation, secured with an API key validation middleware.\n- Updated configuration to include AdminApiEnabled and AdminApiKey settings.\n\n* refactor: update plugin reinstallation endpoint and improve unauthorized response\n\n- Changed the plugin reinstallation endpoint from \"/plugins/reinstall\" to \"/plugin/serverless/reinstall\" for better clarity.\n- Modified the unauthorized response in the AdminAPIKey middleware to return a more descriptive JSON message.",
          "timestamp": "2025-05-15T15:22:25+08:00",
          "tree_id": "c0db8072e3ec38f9edb9b6da8e46a75be9bb2ced",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d884cab8aec86ab9a27c5c3ca1ebd5686c8e3db7"
        },
        "date": 1747293983258,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139910451,
            "unit": "ns/op\t 1593534 B/op\t   28316 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139910451,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593534,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.58,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "978430635 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.58,
            "unit": "ns/op",
            "extra": "978430635 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "978430635 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "978430635 times\n4 procs"
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
          "id": "a22b9c17f004e0fb044242a2bc7df67f8ff3f169",
          "message": "feat: add fake OpenAI server for local runtime benchmarking (#286)\n\n- Introduced a new test file to implement a fake OpenAI server that streams responses for benchmarking purposes.\n- Updated the existing benchmark test to utilize the new fake server, enhancing the testing environment for local plugin runtime invocations.\n- Refactored the package name in the benchmark test file for clarity and consistency.",
          "timestamp": "2025-05-15T15:42:43+08:00",
          "tree_id": "67880350c0534a9bce91fb14245a402afcc46a7c",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/a22b9c17f004e0fb044242a2bc7df67f8ff3f169"
        },
        "date": 1747295196431,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138286130,
            "unit": "ns/op\t 1593643 B/op\t   28316 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138286130,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593643,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.25,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "993018589 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.25,
            "unit": "ns/op",
            "extra": "993018589 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "993018589 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "993018589 times\n4 procs"
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
          "id": "31d7c7417ebeb878041e63af46bcb7fed242713a",
          "message": "feat[0.1.0]: introduce 'run' command for local plugin execution (#283)\n\n* feat: introduce 'run' command for local plugin execution\n\n- Added a new command `run` to launch plugins locally, allowing communication through stdin/stdout.\n- Removed the previous `test` command and its associated functionality to streamline the plugin testing process.\n\n* feat: enhance 'run' command with TCP support\n\n- Introduced a new `RunPluginPayload` structure to encapsulate plugin execution parameters.\n- Added TCP communication mode to the `run` command, allowing multiple client connections.\n- Updated command flags to configure run mode and logging options.\n- Implemented client handling and server creation for both stdin/stdout and TCP modes.\n\n* docs: improve comments in RunPlugin function for clarity\n\n- Enhanced comments to provide clearer explanations of the plugin decoding process and the creation of client streams for both stdin/stdout and TCP modes.\n- Updated comments to reflect the functionality and behavior of the plugin execution flow.\n\n* refactor: update command structure and enhance plugin invocation handling\n\n- Changed the command structure to add `runPluginCommand` under `pluginCommand` for better organization.\n- Introduced `InvokePluginPayload` type to encapsulate plugin invocation details.\n- Enhanced `RunPlugin` function to handle responses and errors more effectively, including logging to stdout.\n- Updated client handling to support session management and improved error handling during plugin invocation.\n- Renamed TCP server creation function for consistency.\n\n* feat: enhance plugin response handling and logging capabilities\n\n- Added a new flag `--response-format` to specify the output format (text or json) for plugin responses.\n- Introduced a `logger` to manage logging output to stdout with timestamps and file information.\n- Updated `logResponse` and `systemLog` functions to handle different response formats.\n- Enhanced `handleClient` and `RunPlugin` functions to utilize the new response format feature.\n- Implemented signal handling to clean up temporary directories on shutdown.\n\n* feat: enhance plugin response structure and logging\n\n- Added `InvokeID` to `InvokePluginPayload` and `GenericResponse` for better tracking of plugin invocations.\n- Updated `logResponse` to include `InvokeID` in error responses for improved debugging.\n- Enhanced client handling in `handleClient` to log plugin readiness and received requests.\n- Refactored client stream creation for better readability and consistency.\n\n* feat: add plugin invoke end response type and logging\n\n- Introduced `GENERIC_RESPONSE_TYPE_PLUGIN_INVOKE_END` to enhance response tracking for plugin invocations.\n- Updated `handleClient` to log the end of plugin invocation, improving visibility into the plugin lifecycle.\n\n* chore: remove fullfeature tags",
          "timestamp": "2025-05-16T14:20:29+08:00",
          "tree_id": "f833b2ce84ada4d54052b434824dd87c8b7286cc",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/31d7c7417ebeb878041e63af46bcb7fed242713a"
        },
        "date": 1747376656380,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139743469,
            "unit": "ns/op\t 1594128 B/op\t   28318 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139743469,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1594128,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28318,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.47,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "974906042 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.47,
            "unit": "ns/op",
            "extra": "974906042 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "974906042 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "974906042 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "62681807+bravomark@users.noreply.github.com",
            "name": "bravomark",
            "username": "bravomark"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "458542ca9fd325337475c553a34f953b0d0e8716",
          "message": "fix: Fix the issue where the List method of Alibaba Cloud OSS did not correctly return plugin file paths. (#287)",
          "timestamp": "2025-05-20T14:11:57+08:00",
          "tree_id": "5c34dea98de36875ab9327434cd93aaf863fa205",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/458542ca9fd325337475c553a34f953b0d0e8716"
        },
        "date": 1747721753278,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141360653,
            "unit": "ns/op\t 1593336 B/op\t   28317 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141360653,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593336,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.52,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "994082896 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.52,
            "unit": "ns/op",
            "extra": "994082896 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "994082896 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "994082896 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "erigocn@gmail.com",
            "name": "Zhi",
            "username": "erigo"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6b112bc8b56ffaef3625f1098acaf067204bcc8a",
          "message": "feat(redis): Add support for Redis Sentinel mode (#276)\n\n* feat(redis): Add support for Redis Sentinel mode\n\nAdded support for Redis Sentinel mode to the Redis client, enabling automatic discovery and connection to the primary node through Sentinel. Updated relevant configuration files and initialization logic to support Sentinel mode configuration and connection.\n\n* add lost RedisUser.",
          "timestamp": "2025-05-20T14:23:01+08:00",
          "tree_id": "52f4f6f60c8a52a509af2265c1d7674bdf0ea07e",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6b112bc8b56ffaef3625f1098acaf067204bcc8a"
        },
        "date": 1747722353606,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139413313,
            "unit": "ns/op\t 1593537 B/op\t   28317 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139413313,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593537,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.95,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "989552910 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.95,
            "unit": "ns/op",
            "extra": "989552910 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "989552910 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "989552910 times\n4 procs"
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
          "id": "c799d90e884dc73cb62fecd0df901380c224d14c",
          "message": "chore: coding style (#291)\n\n- Renamed parameters for consistency and clarity, changing `tenant_id`, `plugin_unique_identifier`, and `install_type` to `tenantId`, `pluginUniqueIdentifier`, and `installType` respectively across multiple functions.\n- Updated corresponding database queries to reflect the new parameter names, enhancing code readability and maintainability.",
          "timestamp": "2025-05-20T17:14:44+08:00",
          "tree_id": "d2d740a6ee22dc3a7453ded8d258f27ea022ff36",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/c799d90e884dc73cb62fecd0df901380c224d14c"
        },
        "date": 1747732658537,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141348315,
            "unit": "ns/op\t 1593802 B/op\t   28317 allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141348315,
            "unit": "ns/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593802,
            "unit": "B/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.52,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "987837642 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.52,
            "unit": "ns/op",
            "extra": "987837642 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "987837642 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "987837642 times\n4 procs"
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
          "id": "9a1da25d5901b22a79502bce6c252a7350dc4bd0",
          "message": "feat: Enhance plugin signing with authorized category verification (#293)\n\n* feat: Enhance plugin signing with authorized category verification\n\n- Added support for an `authorized_category` flag in the signature command to validate the category before signing.\n- Updated the `Sign` function to accept a verification parameter, allowing for category-based signing.\n- Enhanced error handling for invalid categories during the signing process.\n- Updated tests to cover new verification scenarios and ensure proper functionality with the authorized category.\n\n* fix\n\n* fix\n\n* test\n\n* test: Add unit test for plugin verification without verification field\n\n- Introduced a new test case to verify the behavior of plugins that lack a verification field.\n- Updated the signature_test.go file to include the test, ensuring proper functionality of the signing process.\n- Removed the outdated verifier_test.go file and associated test data to streamline the codebase.",
          "timestamp": "2025-05-21T20:05:45+08:00",
          "tree_id": "777d56028899195576ddd65c6f7459404b87efe6",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9a1da25d5901b22a79502bce6c252a7350dc4bd0"
        },
        "date": 1747829318508,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 144759148,
            "unit": "ns/op\t 1594139 B/op\t   28318 allocs/op",
            "extra": "238 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 144759148,
            "unit": "ns/op",
            "extra": "238 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1594139,
            "unit": "B/op",
            "extra": "238 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28318,
            "unit": "allocs/op",
            "extra": "238 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.53,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "986140659 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.53,
            "unit": "ns/op",
            "extra": "986140659 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "986140659 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "986140659 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "altair96wby@gmail.com",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "cdf3493c353738a4c497219e1fe54dac4843fc4b",
          "message": "Update issue templates",
          "timestamp": "2025-05-22T16:46:13+08:00",
          "tree_id": "5fecbd502ba4f2a7c3e0b71244c039182131c6d6",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/cdf3493c353738a4c497219e1fe54dac4843fc4b"
        },
        "date": 1747903746290,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142651575,
            "unit": "ns/op\t 1593977 B/op\t   28318 allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142651575,
            "unit": "ns/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593977,
            "unit": "B/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28318,
            "unit": "allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.17,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.17,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "8380c1d6faa8976acab08269093beb505f036a73",
          "message": "fix(lock): Add concurrency test for Redis lock functionality (#305)\n\n- Introduced a new test case `TestLock` to validate the behavior of the Redis locking mechanism under concurrent access.\n- Enhanced the `Lock` function to improve error handling and ensure proper locking behavior.\n- Utilized `sync.WaitGroup` and atomic operations to measure wait times during lock acquisition, ensuring the lock behaves as expected under high concurrency.",
          "timestamp": "2025-05-23T14:39:37+08:00",
          "tree_id": "d32f48184ab1301d539f3fffc2e39a22a13bcf47",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/8380c1d6faa8976acab08269093beb505f036a73"
        },
        "date": 1747982553714,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139466199,
            "unit": "ns/op\t 1593296 B/op\t   28317 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139466199,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593296,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.74,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "977214750 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.74,
            "unit": "ns/op",
            "extra": "977214750 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "977214750 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "977214750 times\n4 procs"
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
          "id": "3d28e0ceed2bcca29dd5e92596c906f71fbb995a",
          "message": "feat: Add code generation for plugin controllers and services (#301)\n\n* feat: Add code generation for plugin controllers and services\n\n- Introduced a code generation mechanism for plugin controllers and services, allowing for automatic generation based on defined dispatchers.\n- Created new files for generated controllers, services, and templates to streamline the plugin invocation process.\n- Removed outdated functions related to tool validation and runtime parameters, consolidating functionality into generated files.\n- Updated dependencies in go.mod and go.sum to include necessary packages for the new code generation features.\n\n* fix",
          "timestamp": "2025-05-23T14:57:56+08:00",
          "tree_id": "8758f4b182d8f18b81d0015f47c754068de5dbfe",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/3d28e0ceed2bcca29dd5e92596c906f71fbb995a"
        },
        "date": 1747983714286,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139162389,
            "unit": "ns/op\t 1593493 B/op\t   28318 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139162389,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593493,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28318,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.97,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "988813171 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.97,
            "unit": "ns/op",
            "extra": "988813171 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "988813171 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "988813171 times\n4 procs"
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
          "id": "b6906f7eb5c3b311f4f29215d625f32e829350b4",
          "message": "feat: Generate HTTP server routes from template (#306)\n\n* feat: Generate HTTP server routes from template\n\n- Added a new file `http_server.gen.go` to automatically generate HTTP server routes based on defined dispatchers.\n- Refactored existing route definitions in `http_server.go` to utilize the generated routes, improving maintainability.\n- Introduced a code generation function in `generator.go` to create the HTTP server file, enhancing the plugin development workflow.\n- Updated the template for HTTP server generation to streamline route creation for various controllers.\n\n* fix: Update OAuth paths in PluginDispatchers for consistency\n\n- Changed the path for authorization URL from `/oauth/authorization_url` to `/oauth/get_authorization_url`.\n- Updated the path for credentials from `/oauth/credentials` to `/oauth/get_credentials` to align with naming conventions.",
          "timestamp": "2025-05-23T15:27:43+08:00",
          "tree_id": "81a21c026e60a66e16272c6da05e27cae94024da",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b6906f7eb5c3b311f4f29215d625f32e829350b4"
        },
        "date": 1747985426000,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139205345,
            "unit": "ns/op\t 1593467 B/op\t   28317 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139205345,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593467,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.01,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984414014 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.01,
            "unit": "ns/op",
            "extra": "984414014 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984414014 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984414014 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "15627489+NeatGuyCoding@users.noreply.github.com",
            "name": "NeatGuyCoding",
            "username": "NeatGuyCoding"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6b7172d6a60ac4aa6c4d7ca1ab191dca930ab2ec",
          "message": "fix: errChan failed to write response because of panic nil (#296) (#297)\n\n* fix: errChan failed to write response because of panic nil (#296)\n\n* fix: join err and er into a single error using errors.Join, thanks @Yeuoly (#296)\n\n---------\n\nCo-authored-by: NeatGuyCoding <cto@sb>",
          "timestamp": "2025-05-26T11:25:11+08:00",
          "tree_id": "ed39b6d007bc6f3699ea5ac3eaa47e645290d89f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6b7172d6a60ac4aa6c4d7ca1ab191dca930ab2ec"
        },
        "date": 1748230081161,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137694237,
            "unit": "ns/op\t 1593097 B/op\t   28317 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137694237,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593097,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.25,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "964228320 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.25,
            "unit": "ns/op",
            "extra": "964228320 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "964228320 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "964228320 times\n4 procs"
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
          "id": "478c98da5c7d44108ef98c80e44eebf5a93944d8",
          "message": "fix: signature dose not work as expected, if upload new pkg to old dify (#311)\n\n- Updated the  method in the  interface to remove the  parameter, simplifying its usage.\n- Introduced a new  function to provide a default verification structure.\n- Added a  file to store verification data, improving the plugin signing process.\n- Enhanced tests in  to validate the verification process, ensuring proper handling of success and failure scenarios.\n- Refactored related code to accommodate the new verification structure and improve overall maintainability.",
          "timestamp": "2025-05-26T13:10:34+08:00",
          "tree_id": "8c5bff5c8cc3cf163e0b4ae843164491d8585aeb",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/478c98da5c7d44108ef98c80e44eebf5a93944d8"
        },
        "date": 1748236422389,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139348281,
            "unit": "ns/op\t 1593558 B/op\t   28317 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139348281,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593558,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.07,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "974817550 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.07,
            "unit": "ns/op",
            "extra": "974817550 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "974817550 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "974817550 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "slm_1990@126.com",
            "name": "Good Wood",
            "username": "LeeeeeeM"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2cd64adf3b96c866a0f27d23cd784a2f2d38ab9e",
          "message": "feat: change listPlugin struct & add total (#302)",
          "timestamp": "2025-05-27T12:53:37+08:00",
          "tree_id": "ec0a9fab620ce34d05269ae04ee76a1a9d6900f5",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/2cd64adf3b96c866a0f27d23cd784a2f2d38ab9e"
        },
        "date": 1748321790547,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 144909005,
            "unit": "ns/op\t 1593885 B/op\t   28317 allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 144909005,
            "unit": "ns/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593885,
            "unit": "B/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.95,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "970114101 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.95,
            "unit": "ns/op",
            "extra": "970114101 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "970114101 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "970114101 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b3c68cbeecb7efc4b84ebc2dab481cb62614f85e",
          "message": "add packaged file info when plugin package larger than max size (#312)",
          "timestamp": "2025-05-27T13:02:25+08:00",
          "tree_id": "1d043ea91c0d5edd246607972af5512661715b67",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b3c68cbeecb7efc4b84ebc2dab481cb62614f85e"
        },
        "date": 1748322311944,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139072954,
            "unit": "ns/op\t 1593792 B/op\t   28317 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139072954,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593792,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.23,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "986288181 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.23,
            "unit": "ns/op",
            "extra": "986288181 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "986288181 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "986288181 times\n4 procs"
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
          "id": "3918b377f2888f772b7a86f2e32d019eea4625df",
          "message": "refactor: streamline plugin initialization and update YAML templates for consistency (#313)\n\n- Removed redundant flag retrieval in the plugin initialization process, simplifying the code.\n- Updated permission handling to use a single `permissionRequirement` structure for better clarity and maintainability.\n- Enhanced YAML templates by adding quotes around dynamic values to ensure proper formatting and prevent potential parsing issues.",
          "timestamp": "2025-05-27T19:48:11+08:00",
          "tree_id": "1c3420182b9385b5233f1f082da2a4a8622ab1a5",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/3918b377f2888f772b7a86f2e32d019eea4625df"
        },
        "date": 1748346659470,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137211823,
            "unit": "ns/op\t 1593935 B/op\t   28317 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137211823,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593935,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.96,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "999987208 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.96,
            "unit": "ns/op",
            "extra": "999987208 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "999987208 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "999987208 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron Wang",
            "username": "41tair"
          },
          "committer": {
            "email": "byron@dify.ai",
            "name": "Byron Wang",
            "username": "41tair"
          },
          "distinct": true,
          "id": "1fb2d1b532ea0c73280651169c8a3a781a612dad",
          "message": "update issute template: add self checks",
          "timestamp": "2025-05-28T19:04:13+08:00",
          "tree_id": "bfb334d6753b654aaab695cdc6543a891731bed9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1fb2d1b532ea0c73280651169c8a3a781a612dad"
        },
        "date": 1748430512311,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140168512,
            "unit": "ns/op\t 1593528 B/op\t   28317 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140168512,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593528,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.66,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981413559 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.66,
            "unit": "ns/op",
            "extra": "981413559 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981413559 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981413559 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "novice12185727@gmail.com",
            "name": "Novice",
            "username": "Nov1c444"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "0167554a0379d177d29eb1775e41e50b955853db",
          "message": "feat: add mcp tool type (#315)",
          "timestamp": "2025-05-30T14:05:20+08:00",
          "tree_id": "a9120464f848f0c364f1a7f3d5c88937495f9b04",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/0167554a0379d177d29eb1775e41e50b955853db"
        },
        "date": 1748585349047,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138497622,
            "unit": "ns/op\t 1593407 B/op\t   28317 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138497622,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1593407,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28317,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.66,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "966556731 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.66,
            "unit": "ns/op",
            "extra": "966556731 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "966556731 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "966556731 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1c9e28bc7543143a77e9b7e21e28a9d57f984b62",
          "message": "Feat: Replace the internal/oss module with dify-cloud-kit (#317)\n\n* replace internal oss with dify-cloud-kit\n\n* remove validate\n\n* fix tests\n\n* fix tests",
          "timestamp": "2025-05-30T16:44:59+08:00",
          "tree_id": "095c00450e2febc83bbc2bce5a7577fd76ee47da",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1c9e28bc7543143a77e9b7e21e28a9d57f984b62"
        },
        "date": 1748594937686,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139415957,
            "unit": "ns/op\t 1591451 B/op\t   28315 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139415957,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591451,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28315,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.48,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981895969 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.48,
            "unit": "ns/op",
            "extra": "981895969 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981895969 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981895969 times\n4 procs"
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
          "id": "f8914412d9a55dcc55f8130da85d9318617d222a",
          "message": "fix: support serverless plugin management with execution timeout (#318)\n\n- Added `pluginMaxExecutionTimeout` to `PluginManager` for configurable execution limits.\n- Updated `ServerlessPluginRuntime` to utilize the new timeout setting in HTTP requests.\n- Refactored AWSPluginRuntime references to ServerlessPluginRuntime for consistency across the codebase.",
          "timestamp": "2025-05-30T18:08:24+08:00",
          "tree_id": "f78751f1ff86f08c5fdcfe21c384901d64a85ac6",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/f8914412d9a55dcc55f8130da85d9318617d222a"
        },
        "date": 1748599875538,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141204710,
            "unit": "ns/op\t 1591796 B/op\t   28315 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141204710,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591796,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28315,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.84,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "977488501 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.84,
            "unit": "ns/op",
            "extra": "977488501 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "977488501 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "977488501 times\n4 procs"
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
          "id": "052cd0c410873ac879f5f59119b545d2fff2f7d6",
          "message": "fix: remove redundant Content-Type header for payloadReader in HTTP request builder (#320)",
          "timestamp": "2025-05-30T18:49:19+08:00",
          "tree_id": "6a93c6b738ee129831471002e494902fdbc107d9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/052cd0c410873ac879f5f59119b545d2fff2f7d6"
        },
        "date": 1748602334305,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142701271,
            "unit": "ns/op\t 1591773 B/op\t   28315 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142701271,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591773,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28315,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.31,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "993598309 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.31,
            "unit": "ns/op",
            "extra": "993598309 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "993598309 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "993598309 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "zhanluxianshen@163.com",
            "name": "湛露先生",
            "username": "zhanluxianshen"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "5573e1fbd11b322f3903d80949f8e1f88a0b5d8b",
          "message": "Fix env read bug for GCS_CREDENTIALS. (#324)",
          "timestamp": "2025-06-02T09:03:30+08:00",
          "tree_id": "03f3742c5dfa4acdddcd2feb197dcb193d37b10d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/5573e1fbd11b322f3903d80949f8e1f88a0b5d8b"
        },
        "date": 1748826387293,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140945358,
            "unit": "ns/op\t 1591875 B/op\t   28315 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140945358,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591875,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28315,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.18,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "996693678 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.18,
            "unit": "ns/op",
            "extra": "996693678 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "996693678 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "996693678 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "zhanluxianshen@163.com",
            "name": "湛露先生",
            "username": "zhanluxianshen"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "412084b1d88b0ff02499f352168759dc1971336b",
          "message": "fix build error, go.mod upgrade for github.com/panjf2000/ants/v2 (#323)\n\nSigned-off-by: zhanluxianshen <zhanluxianshen@163.com>",
          "timestamp": "2025-06-03T16:17:11+08:00",
          "tree_id": "02d507d3f8ec3cabeeff31577ec839a2af9f48b6",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/412084b1d88b0ff02499f352168759dc1971336b"
        },
        "date": 1748938868875,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140576860,
            "unit": "ns/op\t 1591904 B/op\t   28316 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140576860,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591904,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.43,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982514518 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.43,
            "unit": "ns/op",
            "extra": "982514518 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982514518 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982514518 times\n4 procs"
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
          "id": "5f8072c9823999d40f902ac8def85f1fa079d9c2",
          "message": "Chore/unify configurations (#319)\n\n* refactor: update PluginManager to use configuration for various configurations\n\n- Replaced hardcoded values in PluginManager methods with values from the configuration.\n- Updated serverless plugin launch timeout and working paths to utilize the new configuration structure.\n- Enhanced local plugin runtime initialization to pull settings from the configuration, improving maintainability and flexibility.\n\n* refactor: clean up PluginManager by removing unused fields and updating platform check\n\n- Removed commented-out fields from PluginManager to enhance code clarity.\n- Updated platform check to utilize the configuration structure instead of a direct field reference, improving maintainability.",
          "timestamp": "2025-06-04T20:18:13+08:00",
          "tree_id": "671f26ca39ce507c73db61b3c13add5652c3ccf4",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/5f8072c9823999d40f902ac8def85f1fa079d9c2"
        },
        "date": 1749039663025,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138256186,
            "unit": "ns/op\t 1592000 B/op\t   28316 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138256186,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1592000,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.32,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "973795165 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.32,
            "unit": "ns/op",
            "extra": "973795165 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "973795165 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "973795165 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "novice12185727@gmail.com",
            "name": "Novice",
            "username": "Nov1c444"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8c9458cf2f2ee09ce1c8d3e5bebc1dc30ee82703",
          "message": "Merge pull request #325 from Nov1c444/feat/add-meta\n\nfeat: agent plugin add meta version",
          "timestamp": "2025-06-05T09:44:25+08:00",
          "tree_id": "d7e03e8a1f98563481092f3a2456c188d93b5d6f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/8c9458cf2f2ee09ce1c8d3e5bebc1dc30ee82703"
        },
        "date": 1749088032247,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143526352,
            "unit": "ns/op\t 1592176 B/op\t   28316 allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143526352,
            "unit": "ns/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1592176,
            "unit": "B/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.25,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "957185995 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.25,
            "unit": "ns/op",
            "extra": "957185995 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "957185995 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "957185995 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6873c3f60005cc448ec05b489157ccfe4d1e1427",
          "message": "bump dify-cloud-kit version to 681efb7762a4 (#339)",
          "timestamp": "2025-06-10T16:48:00+08:00",
          "tree_id": "23a3ff7557426aafa7a416401e8e9d5680f3fcf9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6873c3f60005cc448ec05b489157ccfe4d1e1427"
        },
        "date": 1749545510459,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142693944,
            "unit": "ns/op\t 1591775 B/op\t   28316 allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142693944,
            "unit": "ns/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591775,
            "unit": "B/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.25,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997863291 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.25,
            "unit": "ns/op",
            "extra": "997863291 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997863291 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997863291 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7c1e46f643dc38a105a9ab3c90e73f8489daee1d",
          "message": "add serverless runtime interface docs  (#338)\n\n* add sri docs\n\n* add refer to readme\n\n* format readme",
          "timestamp": "2025-06-10T16:48:09+08:00",
          "tree_id": "602b9a9a40704f059c6d59df1b29ec61ee15d767",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7c1e46f643dc38a105a9ab3c90e73f8489daee1d"
        },
        "date": 1749545525285,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143516694,
            "unit": "ns/op\t 1591802 B/op\t   28316 allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143516694,
            "unit": "ns/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591802,
            "unit": "B/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.79,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "971143624 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.79,
            "unit": "ns/op",
            "extra": "971143624 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "971143624 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "971143624 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "debb3744c0fbb424b743977ce63193d9c1ad1b2f",
          "message": "add USE_AWS_S3 args avoid ambiguity. (#340)",
          "timestamp": "2025-06-11T14:32:53+08:00",
          "tree_id": "e92f9c0e14d2bf991a84d3437c39a291fe7f8471",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/debb3744c0fbb424b743977ce63193d9c1ad1b2f"
        },
        "date": 1749623806574,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141414359,
            "unit": "ns/op\t 1591409 B/op\t   28316 allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141414359,
            "unit": "ns/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591409,
            "unit": "B/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.71,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969567933 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.71,
            "unit": "ns/op",
            "extra": "969567933 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969567933 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969567933 times\n4 procs"
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
          "id": "ac6441746d26d3e0c3a0e75800ea51639296859c",
          "message": "feat: add length-prefixed HTTP chunking functionality (#341)\n\n- Introduced new constants for HTTP option types to improve code readability and maintainability.\n- Updated existing HTTP option functions to utilize the new constants.\n- Implemented line-based and length-prefixed chunking methods for improved data processing in HTTP requests.\n- Added comprehensive tests for chunking functionality to ensure reliability and correctness.",
          "timestamp": "2025-06-11T14:45:21+08:00",
          "tree_id": "e6824205a9319b943715dce540f66b7f3ef3daf5",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ac6441746d26d3e0c3a0e75800ea51639296859c"
        },
        "date": 1749624495841,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143563458,
            "unit": "ns/op\t 1591914 B/op\t   28316 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143563458,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591914,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.36,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "991475038 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.36,
            "unit": "ns/op",
            "extra": "991475038 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "991475038 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "991475038 times\n4 procs"
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
          "id": "1e260eedf4f07b81a47f79228604a88d58b4da2d",
          "message": "refactor: using length-prefixed chunking for Backwards invocations (#342)\n\n- Adjusted the header byte manipulation in chunking functions to correctly use the first four bytes for data length.\n- Modified the HTTP request streaming function to include the length-prefixed option for improved data handling.",
          "timestamp": "2025-06-11T16:01:53+08:00",
          "tree_id": "15aa4eb112025f3d803074e80bb22512eb94a85c",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1e260eedf4f07b81a47f79228604a88d58b4da2d"
        },
        "date": 1749629088152,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137274710,
            "unit": "ns/op\t 1591120 B/op\t   28315 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137274710,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1591120,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28315,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.93,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "988449916 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.93,
            "unit": "ns/op",
            "extra": "988449916 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "988449916 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "988449916 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "18e91bbb37834a0e924926d02fbab096d8b7bd0d",
          "message": "fix s3 client path style not used (#344)\n\n* bump cloud-kit version to fix #343\n\n* change env name USE_AWS_S3 to S3_USE_AWS\n\n* update s3_use_aws default value to true\n\n* update readme about the upgrade notice",
          "timestamp": "2025-06-13T15:10:21+08:00",
          "tree_id": "9f4f8621e2ca753639b5b4dfc57e0b8525b5d212",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/18e91bbb37834a0e924926d02fbab096d8b7bd0d"
        },
        "date": 1749798856382,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138662448,
            "unit": "ns/op\t 1592104 B/op\t   28316 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138662448,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1592104,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 28316,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.52,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "967655902 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.52,
            "unit": "ns/op",
            "extra": "967655902 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "967655902 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "967655902 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "guanz42@163.com",
            "name": "Nevermore",
            "username": "guanz42"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6cea2d401e1d7c08572918e5b79fa0eda6354e51",
          "message": "refactor(local_runtime): optimize listener lookup in stdioHolder (#345)\n\nDirectly fetch and invoke the listener for a given session_id instead of iterating over the entire listener map.\n\nSigned-off-by: guanz42 <guanz42@163.com>",
          "timestamp": "2025-06-16T18:31:46+08:00",
          "tree_id": "82904c1ea38ef7a89ad12183badd4c28c1ccf716",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6cea2d401e1d7c08572918e5b79fa0eda6354e51"
        },
        "date": 1750070067301,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139767604,
            "unit": "ns/op\t 1569512 B/op\t   27514 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139767604,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569512,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.88,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "988686522 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.88,
            "unit": "ns/op",
            "extra": "988686522 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "988686522 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "988686522 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "nghuutho74@gmail.com",
            "name": "Rhys",
            "username": "nht1206"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3d1e2aba0e7014a9c001969fa82623a4f571dc6e",
          "message": "fix: skip waiting if error occured (#337)\n\n* fix: skip waiting if error occured\n\n* Update internal/core/plugin_manager/watcher.go\n\nCo-authored-by: Rhys <nghuutho74@gmail.com>\n\n* fix: handle nil error channels in plugin manager\n\n---------\n\nCo-authored-by: Yeuoly <45712896+Yeuoly@users.noreply.github.com>\nCo-authored-by: Yeuoly <admin@srmxy.cn>",
          "timestamp": "2025-06-16T19:15:15+08:00",
          "tree_id": "af7e5b516702df49c17b71fa9a8c07479a47e8f1",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/3d1e2aba0e7014a9c001969fa82623a4f571dc6e"
        },
        "date": 1750072684893,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141008034,
            "unit": "ns/op\t 1569164 B/op\t   27514 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141008034,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569164,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.18,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "976018207 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.18,
            "unit": "ns/op",
            "extra": "976018207 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "976018207 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "976018207 times\n4 procs"
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
          "id": "dfc9622e9117a372f4e8f7de4fed39761ab82cbc",
          "message": "feat(db): enhance database configuration with charset and extras support (#347)\n\n- Updated .env.example to include DB_EXTRAS and DB_CHARSET variables.\n- Refactored InitPluginDB functions for PostgreSQL and MySQL to accept a configuration struct, allowing for more flexible database connection settings.\n- Adjusted connection pool settings to utilize new configuration options for charset and extras.\n\nThis change improves the configurability of database connections and prepares the codebase for future enhancements.",
          "timestamp": "2025-06-18T12:08:56+08:00",
          "tree_id": "d57b51968d946b517eadf549a162b6dd1ee1ce77",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/dfc9622e9117a372f4e8f7de4fed39761ab82cbc"
        },
        "date": 1750219906139,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140116664,
            "unit": "ns/op\t 1569247 B/op\t   27514 allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140116664,
            "unit": "ns/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569247,
            "unit": "B/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.52,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "961885954 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.52,
            "unit": "ns/op",
            "extra": "961885954 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "961885954 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "961885954 times\n4 procs"
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
          "id": "a6c8fae9c7a81deeb6e8eff12eedb01383489862",
          "message": "feat: add decode plugin from identifier endpoint (#349)\n\n* feat: add decode plugin from identifier endpoint\n\n- Introduced a new endpoint to decode a plugin from a unique identifier.\n- Implemented the DecodePluginFromIdentifier function to handle decoding and verification of plugin signatures.\n- Updated the HTTP server routes to include the new decode endpoint.\n\n* refactor: update decode plugin endpoint path\n\n- Moved the decode plugin from identifier endpoint to a new path for better organization.\n- Updated the HTTP server routes accordingly to reflect the new endpoint structure.",
          "timestamp": "2025-06-18T16:01:15+08:00",
          "tree_id": "093c544dd2a09775b758457aaf57db72271d102f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/a6c8fae9c7a81deeb6e8eff12eedb01383489862"
        },
        "date": 1750233841697,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139248827,
            "unit": "ns/op\t 1569166 B/op\t   27513 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139248827,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569166,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27513,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.98,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.98,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "364776488@qq.com",
            "name": "Ganondorf",
            "username": "ZombieBlue"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ff875c7e7a63e91315413b2fb1f97a83725a0a15",
          "message": "Split REMOTE_INSTALL_ADDRESS into HOST and PORT in .env.example to align with the official docs (#356)\n\nCo-authored-by: lizb <lizb@sugon.com>",
          "timestamp": "2025-06-20T10:33:28+08:00",
          "tree_id": "090006e02bfad4d15740cb3e2916453e5c023654",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ff875c7e7a63e91315413b2fb1f97a83725a0a15"
        },
        "date": 1750387039494,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140961343,
            "unit": "ns/op\t 1569499 B/op\t   27514 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140961343,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569499,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.2,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "988069286 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.2,
            "unit": "ns/op",
            "extra": "988069286 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "988069286 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "988069286 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "tsonglew@gmail.com",
            "name": "Tsonglew",
            "username": "tsonglew"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "166609f3f6a4f67f4fc74f411acdc237ec29093c",
          "message": "fix: launch error when using redis sentinel (#352)",
          "timestamp": "2025-06-23T16:51:22+08:00",
          "tree_id": "3bba7170ade46e21e540b67a1880f22328c1abc4",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/166609f3f6a4f67f4fc74f411acdc237ec29093c"
        },
        "date": 1750668907792,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139252092,
            "unit": "ns/op\t 1569375 B/op\t   27514 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139252092,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569375,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.67,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969203689 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.67,
            "unit": "ns/op",
            "extra": "969203689 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969203689 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969203689 times\n4 procs"
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
          "id": "ae2658dd43ccad8dafd24ca0ba6ef0bedfdf1be1",
          "message": "refactor(plugin_manager): remove first logging of local plugin launch (#357)",
          "timestamp": "2025-06-24T13:17:19+08:00",
          "tree_id": "81b37c1a6b85410c6aefa5cc03c459b0f3f1c1b9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ae2658dd43ccad8dafd24ca0ba6ef0bedfdf1be1"
        },
        "date": 1750742474090,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 144121750,
            "unit": "ns/op\t 1569379 B/op\t   27514 allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 144121750,
            "unit": "ns/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569379,
            "unit": "B/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.78,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "968670078 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.78,
            "unit": "ns/op",
            "extra": "968670078 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "968670078 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "968670078 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9c6bbc687f98426a2c599e3d80e7811ec4a1eec1",
          "message": "refactor: extract DSN construction to buildDSN for better reuse and readability (#360)",
          "timestamp": "2025-06-25T14:15:47+08:00",
          "tree_id": "5bcadff2de24f2d7172af8c40141a38f4ba7a668",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9c6bbc687f98426a2c599e3d80e7811ec4a1eec1"
        },
        "date": 1750832377800,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141446819,
            "unit": "ns/op\t 1569184 B/op\t   27514 allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141446819,
            "unit": "ns/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569184,
            "unit": "B/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.25,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "987138574 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.25,
            "unit": "ns/op",
            "extra": "987138574 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "987138574 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "987138574 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "kinoooolu@gmail.com",
            "name": "kinoooolu",
            "username": "kinoooolu"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "f096900e620b0cd623fb4a45b544f4b91453d965",
          "message": "fix:response data will be discard if tool/llm response buffer overflow (#362)\n\nCo-authored-by: kino.lu <kino.lu@vipshop.com>",
          "timestamp": "2025-06-26T11:44:41+08:00",
          "tree_id": "12abc394b7d696ede3a2ca0a52acbb3e72cebaf3",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/f096900e620b0cd623fb4a45b544f4b91453d965"
        },
        "date": 1750909720819,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139208098,
            "unit": "ns/op\t 1569432 B/op\t   27514 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139208098,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569432,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.01,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.01,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "7bb6406894111a071bee0ff8d1d0124fc3037b86",
          "message": "test(stream): add delay in TestStreamCloseBlockingWrite to ensure blocking write completion (#365)",
          "timestamp": "2025-06-26T11:54:54+08:00",
          "tree_id": "850e92f9c37e2b9f3b6307c17bb5b2a81e7f9d10",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7bb6406894111a071bee0ff8d1d0124fc3037b86"
        },
        "date": 1750910326680,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139759057,
            "unit": "ns/op\t 1569587 B/op\t   27514 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139759057,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569587,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.57,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "970923838 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.57,
            "unit": "ns/op",
            "extra": "970923838 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "970923838 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "970923838 times\n4 procs"
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
          "id": "f80d8a8946e6ab7b7b3149aab123a3d2cd2d5638",
          "message": "feat: add pull request template for improved contribution guidelines (#366)\n\n- Introduced a new pull request template to standardize contributions.\n- The template includes sections for description, type of change, essential checklist, and additional information to assist reviewers.",
          "timestamp": "2025-06-26T12:09:55+08:00",
          "tree_id": "68f56717e6031ad4ba748768b55ef8b06c77cf26",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/f80d8a8946e6ab7b7b3149aab123a3d2cd2d5638"
        },
        "date": 1750911162063,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141236840,
            "unit": "ns/op\t 1569511 B/op\t   27514 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141236840,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569511,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.87,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "954974404 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.87,
            "unit": "ns/op",
            "extra": "954974404 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "954974404 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "954974404 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jingfelix@outlook.com",
            "name": "Tianyi Jing",
            "username": "jingfelix"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "af3fec6f327df8192fbfd7ce52b7b8406400a7d2",
          "message": "fix: prevent duplicate packaging (#367)\n\nfixes: https://github.com/langgenius/dify-plugins/issues/612\nfixes: https://github.com/langgenius/dify-plugins/issues/234\n\nSigned-off-by: jingfelix <jingfelix@outlook.com>",
          "timestamp": "2025-06-26T17:45:59+08:00",
          "tree_id": "c61870b8a93b5e5df3486ad6643a29f817a9ced9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/af3fec6f327df8192fbfd7ce52b7b8406400a7d2"
        },
        "date": 1750931322974,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137632589,
            "unit": "ns/op\t 1569203 B/op\t   27514 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137632589,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569203,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.27,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984038859 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.27,
            "unit": "ns/op",
            "extra": "984038859 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984038859 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984038859 times\n4 procs"
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
          "id": "a70d808dd0e8e900d81cb97e0f49254c81711698",
          "message": "feat(dynamic_select): implement dynamic parameter fetching functionality (#358)\n\n* feat(dynamic_select): implement dynamic parameter fetching functionality\n\n- Added FetchDynamicParameterOptions function to handle dynamic parameter selection.\n- Introduced new access type and action for dynamic select in access_types.\n- Updated HTTP server routes to include the new endpoint for fetching dynamic parameters.\n- Created necessary service and controller files for dynamic select operations.\n\n* refactor(access_types): rename dynamic select access type to dynamic parameter\n\n- Updated access type constants to reflect the change from PLUGIN_ACCESS_TYPE_DYNAMIC_SELECT to PLUGIN_ACCESS_TYPE_DYNAMIC_PARAMETER.\n- Adjusted related references in the PluginDispatchers and FetchDynamicParameterOptions function to maintain consistency.",
          "timestamp": "2025-06-27T19:24:59+08:00",
          "tree_id": "1243ad7e37c254e947b195ece9930dd463866a2b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/a70d808dd0e8e900d81cb97e0f49254c81711698"
        },
        "date": 1751023665674,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141363169,
            "unit": "ns/op\t 1569560 B/op\t   27514 allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141363169,
            "unit": "ns/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569560,
            "unit": "B/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.63,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984541032 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.63,
            "unit": "ns/op",
            "extra": "984541032 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984541032 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984541032 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "63482+defstream@users.noreply.github.com",
            "name": "@defstream",
            "username": "defstream"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "7a7848b3aea334615514bbaa5f7455c452f73cf1",
          "message": "Update README.md (#372)",
          "timestamp": "2025-06-28T12:31:57+08:00",
          "tree_id": "299f602726cb6bc67b574066fd4d5e6228fa8606",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7a7848b3aea334615514bbaa5f7455c452f73cf1"
        },
        "date": 1751085287144,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140821066,
            "unit": "ns/op\t 1569241 B/op\t   27514 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140821066,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569241,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.26,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "977814086 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.26,
            "unit": "ns/op",
            "extra": "977814086 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "977814086 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "977814086 times\n4 procs"
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
          "id": "6d6fb3892e4be0682f65bc4f6bc9247d87a4d3dd",
          "message": "feat: add InvokeLLMWithStructuredOutput functionality (#369)\n\n* feat: add InvokeLLMWithStructuredOutput functionality\n\n- Introduced a new method InvokeLLMWithStructuredOutput to the BackwardsInvocation interface for handling structured output requests.\n- Added corresponding request and response types to support structured output.\n- Implemented the method in both RealBackwardsInvocation and MockedDifyInvocation for testing purposes.\n- Updated permission handling and task execution for the new structured output invocation type.\n\nThis enhancement allows for more flexible and detailed responses from the LLM, improving the overall functionality of the invocation system.\n\n* refactor: enhance LLMResultChunkWithStructuredOutput structure\n\n- Updated the LLMResultChunkWithStructuredOutput type to include additional fields: Model, SystemFingerprint, and Delta.\n- Added comments to clarify the reasoning behind the structure and the use of type embedding for JSON marshaling.\n\nThis change improves the clarity and functionality of the LLMResultChunkWithStructuredOutput type, ensuring proper JSON serialization.\n\n* refactor: streamline LLMResultChunk construction in InvokeLLMWithStructuredOutput\n\n- Simplified the construction of LLMResultChunk and LLMResultChunkWithStructuredOutput by removing unnecessary type embedding.\n- Enhanced readability and maintainability of the code while preserving functionality.\n\nThis change contributes to cleaner code and improved clarity in the handling of structured output responses.",
          "timestamp": "2025-06-30T12:07:46+08:00",
          "tree_id": "32821ebca5412671e0c60776f0a631b64b9d3bb8",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6d6fb3892e4be0682f65bc4f6bc9247d87a4d3dd"
        },
        "date": 1751256645263,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143013429,
            "unit": "ns/op\t 1569222 B/op\t   27514 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143013429,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569222,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.11,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "974131809 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.11,
            "unit": "ns/op",
            "extra": "974131809 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "974131809 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "974131809 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "akikevinsama@gmail.com",
            "name": "AkisAya",
            "username": "AkisAya"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "412589f94afa619ed95e1e3ef44ea1a4208693b9",
          "message": "skip error plugin names (#381)",
          "timestamp": "2025-07-03T23:43:03+08:00",
          "tree_id": "6ca4e12bc322ebbcc739f78acb7a3845bfbaf1a8",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/412589f94afa619ed95e1e3ef44ea1a4208693b9"
        },
        "date": 1751557609120,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140332957,
            "unit": "ns/op\t 1569972 B/op\t   27514 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140332957,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569972,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.3,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981172118 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.3,
            "unit": "ns/op",
            "extra": "981172118 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981172118 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981172118 times\n4 procs"
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
          "id": "56fcd686e122c6fded53f45ab73598d58d267368",
          "message": "feat: add active request tracking to health check and dispatch routes (#384)\n\n- Implemented middleware to track active requests and active dispatch requests using atomic counters.\n- Updated health check response to include counts of active requests and active dispatch requests.\n- Integrated the new middleware into the HTTP server and plugin dispatch group for improved monitoring.",
          "timestamp": "2025-07-04T19:56:42+08:00",
          "tree_id": "dfc6aa00b0f98c5646a9c550c6896ef0d3918511",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/56fcd686e122c6fded53f45ab73598d58d267368"
        },
        "date": 1751630429697,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141911834,
            "unit": "ns/op\t 1569709 B/op\t   27514 allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141911834,
            "unit": "ns/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569709,
            "unit": "B/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.03,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "958459635 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.03,
            "unit": "ns/op",
            "extra": "958459635 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "958459635 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "958459635 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "37723456+Blackoutta@users.noreply.github.com",
            "name": "Blackoutta",
            "username": "Blackoutta"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "18d4151883a4e1d0bf10f5ec0918c0853a67bda6",
          "message": "optimize: skip sleep for remote plugin runtime during restart, making the debugging experience smoother (#387)",
          "timestamp": "2025-07-07T18:17:02+08:00",
          "tree_id": "163cef4a859c77f3bfe44e38d014a5c864aaa174",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/18d4151883a4e1d0bf10f5ec0918c0853a67bda6"
        },
        "date": 1751883656563,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139210708,
            "unit": "ns/op\t 1569835 B/op\t   27514 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139210708,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569835,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.33,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "962009116 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.33,
            "unit": "ns/op",
            "extra": "962009116 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "962009116 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "962009116 times\n4 procs"
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
          "id": "b97cce716751cc1185273657c87037dad51e4692",
          "message": "enhance(cli/icon): add multiple categories default plugin icons (#388)\n\n* feat: add support for dark icon\n\n- Introduced IconDark field in PluginDeclaration and related structures to support dark mode icons.\n- Updated the installation process to handle dark icons.\n- Enhanced asset validation to check for the presence of dark icons.\n\nThis change improves the visual consistency of plugins in dark mode environments.\n\n* enhance(cli/icon): add plugin icon support with multiple categories\n\n- Added support for light and dark icons for various plugin categories including agent, datasource, extension, model, tool, and trigger.\n- Replaced the previous single icon implementation with a structured map for better organization and retrieval of icons based on category and theme.\n- Removed the old Python icon file to streamline asset management.\n\nThis update improves the visual representation of plugins across different themes, enhancing user experience.\n\n* change icons\n\n* fix\n\n* fix\n\n* comments",
          "timestamp": "2025-07-08T12:48:48+08:00",
          "tree_id": "e45414b4afe26508e4bded8f56fc4bbcbd54210d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b97cce716751cc1185273657c87037dad51e4692"
        },
        "date": 1751950361118,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139991344,
            "unit": "ns/op\t 1569968 B/op\t   27514 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139991344,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569968,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.93,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "973384834 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.93,
            "unit": "ns/op",
            "extra": "973384834 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "973384834 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "973384834 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "454690042@qq.com",
            "name": "homejim",
            "username": "homejim"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6ae762ba1146695653c18e5534a186d710949e8e",
          "message": "feat(plugin_manager): optimize local plugin startup with concurrency (#375)\n\n* feat(plugin_manager): optimize local plugin startup with concurrent control\n\n- Add semaphore-based concurrency control for plugin launches\n- Implement parallel plugin startup using goroutines\n- Optimize error handling to prevent goroutine blocking\n- Add concurrency metrics logging\n\nNote: handleNewLocalPlugins now accepts config parameter with default concurrency limit\n\n* feat(plugin_manager): make local plugin launching concurrency configurable\n\n* fix(plugin_manager): optimize comment and error handling\n\n- Updated comments to clarify the concurrent plugin launching configuration.\n- Added a nil check for the error channel during plugin startup to improve code robustness.\n\n* refactor(plugin_manager): refactor plugin startup logic\n\n- Remove the semaphore mechanism and switch to using routine.Submit for concurrency management\n\n* fix(plugin_manager): Optimize plugin startup logs and concurrency control\n\n- Added log output for maximum concurrency when starting local plugins\n- Implemented a channel-based concurrency control mechanism to ensure limits are not exceeded\n- Fixed closure variable capture issue to prevent incorrect plugin information\n- Improved error handling to avoid deadlocks during startup\n\n* fix(plugin_manager): simplify error channel handling and semaphore release logic\n\n---------\n\nCo-authored-by: jim02.he <jim02.he@vipshop.com>",
          "timestamp": "2025-07-08T19:09:31+08:00",
          "tree_id": "bd5010065a49c23813ade3b5f01f50b354918933",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6ae762ba1146695653c18e5534a186d710949e8e"
        },
        "date": 1751973205252,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140443350,
            "unit": "ns/op\t 1569418 B/op\t   27514 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140443350,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569418,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.85,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "965073944 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.85,
            "unit": "ns/op",
            "extra": "965073944 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "965073944 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "965073944 times\n4 procs"
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
          "id": "7bc3b7565e89dc52e75cf0650582a9df040ebfd4",
          "message": "feat(plugin_manager): enhance asset remapping for icons (#392)\n\n- Refactored the RemapAssets function to streamline the remapping of icon fields for both models and tools, including support for dark mode icons.\n- Introduced new fields IconSmallDark and IconLargeDark in the ModelProviderDeclaration and added IconDark in ToolProviderIdentity to accommodate dark mode assets.\n- Improved error handling during the remapping process for better clarity and maintainability.",
          "timestamp": "2025-07-09T17:02:57+08:00",
          "tree_id": "351bb1d9c09515f8ef242bc9aa2244a8749181e3",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7bc3b7565e89dc52e75cf0650582a9df040ebfd4"
        },
        "date": 1752052015108,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141015552,
            "unit": "ns/op\t 1569267 B/op\t   27514 allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141015552,
            "unit": "ns/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569267,
            "unit": "B/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "243 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.74,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "963478232 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.74,
            "unit": "ns/op",
            "extra": "963478232 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "963478232 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "963478232 times\n4 procs"
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
          "id": "33c023b37507cfc9f179d4bffbaebf70b7a00664",
          "message": "refactor(plugin_manager): enhance HTTP client timeout handling in ser… (#385)\n\n* refactor(plugin_manager): enhance HTTP client timeout handling in serverless runtime with DialContext\n\n- Updated the HTTP client in the ServerlessPluginRuntime to use a context-aware DialContext for better timeout management.\n- Removed the write timeout option from the HTTP request builder, streamlining the request handling process.\n- Improved connection handling by setting write deadlines based on the PluginMaxExecutionTimeout.\n\nThis change enhances the reliability of network operations within the serverless runtime environment.\n\n* refactor(plugin_manager): remove write deadline setting in serverless runtime connection initialization\n\n- Eliminated the write deadline setting from the connection initialization in the ServerlessPluginRuntime.\n- This change simplifies the connection handling process and aligns with the recent enhancements to timeout management.\n\nThis update contributes to a more streamlined and efficient network operation within the serverless environment.\n\n* refactor(plugin_manager): adjust HTTP client timeout settings in serverless runtime\n\n- Modified the HTTP client configuration in the ServerlessPluginRuntime to set the TLS handshake timeout based on PluginMaxExecutionTimeout.\n- Retained the IdleConnTimeout setting to ensure consistent connection management.\n\nThis update improves the timeout handling for secure connections, enhancing overall network reliability in the serverless environment.",
          "timestamp": "2025-07-09T17:43:24+08:00",
          "tree_id": "62d581add1436cd66db05defd32b43c3633f7197",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/33c023b37507cfc9f179d4bffbaebf70b7a00664"
        },
        "date": 1752054440048,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139927545,
            "unit": "ns/op\t 1568897 B/op\t   27514 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139927545,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1568897,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27514,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.55,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "955520791 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.55,
            "unit": "ns/op",
            "extra": "955520791 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "955520791 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "955520791 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a0414b30e66f33138da75dc4649ef965f7e7a135",
          "message": "0.2.0 (#402)\n\n* feat(oauth): add RedirectURI field to OAuth request structures\n\n* feat(oauth): update OAuthSchema validation\n\n* feat: add Context field to request and session structures\n\n* feat: add CredentialType field to Credentials and InvokeToolRequest structures\n\n* fix: handle unhandled default case in basic_type.go\n\n* feat: add support for build branches in build-push.yml",
          "timestamp": "2025-07-17T16:28:43+08:00",
          "tree_id": "6cde80a358d9c658662bfe3bd9bee06af1c559a7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/a0414b30e66f33138da75dc4649ef965f7e7a135"
        },
        "date": 1752741147394,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139158591,
            "unit": "ns/op\t 1569606 B/op\t   27524 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139158591,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569606,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.51,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "950798575 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.51,
            "unit": "ns/op",
            "extra": "950798575 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "950798575 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "950798575 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "3b0a8679f4461185bbc0f5b5710159500b7f8a8c",
          "message": "feat/tool oauth cli template (#407)\n\n* feat(cli): update OAuth handling and requirements for dify_plugin\n\n* feat(oauth): update OAuth support and adjust dify_plugin version constraints",
          "timestamp": "2025-07-21T15:37:13+08:00",
          "tree_id": "6a0f601078237088c65d0db97c96fdde7f5ebd8d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/3b0a8679f4461185bbc0f5b5710159500b7f8a8c"
        },
        "date": 1753083664913,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140257900,
            "unit": "ns/op\t 1570254 B/op\t   27524 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140257900,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1570254,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.66,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981566634 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.66,
            "unit": "ns/op",
            "extra": "981566634 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981566634 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981566634 times\n4 procs"
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
          "id": "7f463e32f6e1a609f64c0eb49f23b3187055cfdd",
          "message": "feat(plugin_decoder): add support for internationalized readme files (#393)\n\n* feat(plugin_decoder): add support for internationalized readme files\n\n- Introduced the AvailableI18nReadme method in the PluginDecoder interface to retrieve available readme files in multiple languages.\n- Implemented the method in FSPluginDecoder and ZipPluginDecoder to read localized readme files from the filesystem and zip archives.\n- Enhanced UnixPluginDecoder to handle readme files in a structured manner, including support for reading from a dedicated \"readme\" directory.\n- Added unit tests to verify the functionality of the AvailableI18nReadme method and ensure correct retrieval of localized readme content.\n\n* feat(plugin): add support for multilingual README generation\n\n- Introduced functionality to create README files in multiple languages (Simplified Chinese, Japanese, Portuguese) based on user selection.\n- Enhanced the profile management to include options for enabling internationalized README and selecting languages.\n- Added new language choice structure to manage language options and their selection state.\n- Implemented rendering and writing of language-specific README files during plugin creation.\n- Included new README template files for each supported language.\n\n* feat(plugin): add README command and list functionality\n\n- Introduced a new `readme` command to the plugin CLI for managing README files.\n- Added `list` subcommand to display available README languages for a specified plugin path.\n- Implemented functionality to read and list supported README languages in a tabular format.\n- Enhanced error handling for plugin file reading and decoding processes.",
          "timestamp": "2025-07-21T16:02:26+08:00",
          "tree_id": "7c98741d6e8893dd6b6c535348593616f7fbe76b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/7f463e32f6e1a609f64c0eb49f23b3187055cfdd"
        },
        "date": 1753085184731,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137919768,
            "unit": "ns/op\t 1569586 B/op\t   27524 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137919768,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569586,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.41,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981902931 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.41,
            "unit": "ns/op",
            "extra": "981902931 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981902931 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981902931 times\n4 procs"
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
          "id": "bace3bfb580f845f3bbe690d84cda5232a33b3ef",
          "message": "feat(oauth): implement refresh credentials functionality (#408)\n\n* feat(oauth): implement refresh credentials functionality\n\n- Added RefreshCredentials endpoint to handle OAuth credential refresh requests.\n- Introduced RequestOAuthRefreshCredentials structure for request validation.\n- Updated access types and actions to include refresh credentials.\n- Enhanced server routing and controller logic to support the new functionality.\n- Updated OAuth entities to include expiration handling for refreshed credentials.\n\n* feat(oauth): add metadata field to OAuthGetCredentialsResult\n\n---------\n\nCo-authored-by: Harry <xh001x@hotmail.com>",
          "timestamp": "2025-07-23T13:11:36+08:00",
          "tree_id": "480f3aad74cae0b665785c581a281665c0e17e5b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/bace3bfb580f845f3bbe690d84cda5232a33b3ef"
        },
        "date": 1753247736699,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142665650,
            "unit": "ns/op\t 1569593 B/op\t   27524 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142665650,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569593,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.48,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "964223156 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.48,
            "unit": "ns/op",
            "extra": "964223156 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "964223156 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "964223156 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49699333+dependabot[bot]@users.noreply.github.com",
            "name": "dependabot[bot]",
            "username": "dependabot[bot]"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9234aed47984997aa4296207602ee507e4981e31",
          "message": "chore(deps): bump github.com/go-jose/go-jose/v4 from 4.0.4 to 4.0.5 (#409)\n\nBumps [github.com/go-jose/go-jose/v4](https://github.com/go-jose/go-jose) from 4.0.4 to 4.0.5.\n- [Release notes](https://github.com/go-jose/go-jose/releases)\n- [Commits](https://github.com/go-jose/go-jose/compare/v4.0.4...v4.0.5)\n\n---\nupdated-dependencies:\n- dependency-name: github.com/go-jose/go-jose/v4\n  dependency-version: 4.0.5\n  dependency-type: indirect\n...\n\nSigned-off-by: dependabot[bot] <support@github.com>\nCo-authored-by: dependabot[bot] <49699333+dependabot[bot]@users.noreply.github.com>",
          "timestamp": "2025-07-23T13:51:27+08:00",
          "tree_id": "82cc74c96889468da6d8ad8fa58dabc710edfbe0",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9234aed47984997aa4296207602ee507e4981e31"
        },
        "date": 1753250135096,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142824539,
            "unit": "ns/op\t 1570179 B/op\t   27524 allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142824539,
            "unit": "ns/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1570179,
            "unit": "B/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "240 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.96,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "972268657 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.96,
            "unit": "ns/op",
            "extra": "972268657 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "972268657 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "972268657 times\n4 procs"
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
          "id": "80389be55b8b703e07505bbde2dd7154a45802d3",
          "message": "chore(deps): update go-git and golang.org/x/tools dependencies (#410)\n\n- Updated github.com/go-git/go-git from v4.7.0 to v5.16.2.\n- Bumped golang.org/x/tools from v0.22.0 to v0.35.0.\n- Added indirect dependencies for github.com/go-git/gcfg and github.com/go-git/go-billy.\n- Updated various indirect dependencies including golang.org/x/net, golang.org/x/sys, and golang.org/x/text.",
          "timestamp": "2025-07-23T16:22:48+08:00",
          "tree_id": "95b05cfa203f89e30b35430d79710130039673df",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/80389be55b8b703e07505bbde2dd7154a45802d3"
        },
        "date": 1753259206328,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140160859,
            "unit": "ns/op\t 1569775 B/op\t   27524 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140160859,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569775,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.02,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "993541242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.02,
            "unit": "ns/op",
            "extra": "993541242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "993541242 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "993541242 times\n4 procs"
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
          "id": "178e7649ba98f58721db13c9c217707b1281a397",
          "message": "fix(env): update REMOTE_INSTALL_URL format in .env.example (#412)\n\n- Combined REMOTE_INSTALL_URL and REMOTE_INSTALL_PORT into a single line for clarity.\n- This change improves the configuration format for easier understanding and usage.",
          "timestamp": "2025-07-23T18:21:42+08:00",
          "tree_id": "3d3219536253e764c361b2f843f98405433e9494",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/178e7649ba98f58721db13c9c217707b1281a397"
        },
        "date": 1753266275496,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139686616,
            "unit": "ns/op\t 1570095 B/op\t   27524 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139686616,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1570095,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27524,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.15,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "971733394 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.15,
            "unit": "ns/op",
            "extra": "971733394 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "971733394 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "971733394 times\n4 procs"
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
          "id": "222ab80a6b5767327c8e529bb6c73fa56c4273e6",
          "message": "refactor(session_manager): update GetSession function to return error (#413)\n\n- Modified GetSession to return an error alongside the session, improving error handling when a session is not found.\n- Updated invocation of GetSession in AWS event handler to handle the new error return value.\n- Removed logging statements in favor of returning errors directly for better error propagation.",
          "timestamp": "2025-07-24T13:40:38+08:00",
          "tree_id": "bce30900f1221539a631b603a0b2cf00aceae10b",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/222ab80a6b5767327c8e529bb6c73fa56c4273e6"
        },
        "date": 1753335806805,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 135145928,
            "unit": "ns/op\t 1569452 B/op\t   27523 allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 135145928,
            "unit": "ns/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569452,
            "unit": "B/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27523,
            "unit": "allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 35.61,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 35.61,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "N.Betula.Lu@gmail.com",
            "name": "Betula-L",
            "username": "Betula-L"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "21015d447432f633fdf33377e5e7e8f64d635f9f",
          "message": "fix credential type (#419)",
          "timestamp": "2025-08-08T21:56:34+08:00",
          "tree_id": "869a46d8e397dff3a737b66bbbcabf157d247536",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/21015d447432f633fdf33377e5e7e8f64d635f9f"
        },
        "date": 1754661629075,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138242054,
            "unit": "ns/op\t 1569633 B/op\t   27517 allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138242054,
            "unit": "ns/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569633,
            "unit": "B/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "252 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.2,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "992552492 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.2,
            "unit": "ns/op",
            "extra": "992552492 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "992552492 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "992552492 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jingfelix@outlook.com",
            "name": "Tianyi Jing",
            "username": "jingfelix"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "526addc778122f66eeb1c614049aa29d956565ad",
          "message": "fix: add missing types of tool_parameter (#421)\n\nSigned-off-by: jingfelix <jingfelix@outlook.com>",
          "timestamp": "2025-08-23T18:52:32+08:00",
          "tree_id": "0f5685273ebbd26e51f8173026b3ac28638b7b94",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/526addc778122f66eeb1c614049aa29d956565ad"
        },
        "date": 1755946591096,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141602774,
            "unit": "ns/op\t 1569499 B/op\t   27517 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141602774,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569499,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.04,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "975150644 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.04,
            "unit": "ns/op",
            "extra": "975150644 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "975150644 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "975150644 times\n4 procs"
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
          "id": "b286be778ea05f357a1caef27fd2fdb851e36619",
          "message": "chore: remove JSON schema validation from tool and agent services (#430)\n\n* chore: remove JSON schema validation from tool and agent services\n\n- Remove bindAgentStrategyValidator and bindToolValidator functions\n- Remove gojsonschema dependency\n- Simplify InvokeAgentStrategy by removing validation logic\n- Consolidate tool invocation into generated code\n- Remove redundant tool_service.go and related test files\n- Move InvokeTool to generated files for consistency\n\n🤖 Generated with [Claude Code](https://claude.ai/code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>\n\n* fix: remove duplicated routes\n\n---------\n\nCo-authored-by: Claude <noreply@anthropic.com>",
          "timestamp": "2025-08-26T20:36:25+08:00",
          "tree_id": "f9388f9220fb5dfbc661d06ca9a9df8ab9005cb3",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b286be778ea05f357a1caef27fd2fdb851e36619"
        },
        "date": 1756211951574,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139310020,
            "unit": "ns/op\t 1569421 B/op\t   27516 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139310020,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569421,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.26,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997509908 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.26,
            "unit": "ns/op",
            "extra": "997509908 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997509908 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997509908 times\n4 procs"
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
          "id": "ab7540eed990762f297f1c3b33cea847c3ed1dfd",
          "message": "Add Claude Code GitHub Workflow (#434)\n\n* \"Claude PR Assistant workflow\"\n\n* \"Claude Code Review workflow\"",
          "timestamp": "2025-08-27T14:03:59+08:00",
          "tree_id": "34248e0d9d5a3879dbef64847d2a5cea4fbb984d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ab7540eed990762f297f1c3b33cea847c3ed1dfd"
        },
        "date": 1756274813400,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143193683,
            "unit": "ns/op\t 1569699 B/op\t   27517 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143193683,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569699,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.84,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "971012614 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.84,
            "unit": "ns/op",
            "extra": "971012614 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "971012614 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "971012614 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "qkzsky@vip.qq.com",
            "name": "kuangzhiqiang",
            "username": "qkzsky"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "6c03c1b3c87be79a19f8209d91d960ca61f19e07",
          "message": "fix Frequent plugin pulls (#420)\n\nCo-authored-by: kuangzhiqiang <kuangzhiqiang@xiaomi.com>",
          "timestamp": "2025-08-27T14:16:31+08:00",
          "tree_id": "d6c0a75784620d151211ae72fab34431039560a3",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/6c03c1b3c87be79a19f8209d91d960ca61f19e07"
        },
        "date": 1756275563914,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142165801,
            "unit": "ns/op\t 1569507 B/op\t   27516 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142165801,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569507,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.83,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "966277957 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.83,
            "unit": "ns/op",
            "extra": "966277957 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "966277957 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "966277957 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "37723456+Blackoutta@users.noreply.github.com",
            "name": "Blackoutta",
            "username": "Blackoutta"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b86f12fbaac0f70a177dd13ac5a86ae64d09db44",
          "message": "Add Redis caching to FetchPluginInstallation middleware for performance optimization (#400)\n\n* add caching to fetch plugin installation middleware\n\n* invalidate plugin inst cache after successfully upgrading or uninstalling operations\n\n* optimize: refactored cache invalidation ops from curd layer to service layer\n\n---------\n\nCo-authored-by: yangyi.hu593671 <yangyi.hu593671@seres.cn>",
          "timestamp": "2025-08-27T14:16:52+08:00",
          "tree_id": "635630d8b72d2d0bbc409bb1e845f2270bd365b8",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b86f12fbaac0f70a177dd13ac5a86ae64d09db44"
        },
        "date": 1756275575203,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138504316,
            "unit": "ns/op\t 1569367 B/op\t   27517 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138504316,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569367,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.07,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "946082406 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.07,
            "unit": "ns/op",
            "extra": "946082406 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "946082406 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "946082406 times\n4 procs"
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
          "id": "604da5d495eb42280077c9039dc1abbbad46a454",
          "message": "refactor: rename AWS references to Serverless in plugin daemon (#437)\n\n- Updated AWS-related types and functions to reflect Serverless terminology, including renaming `AWSTransactionHandler` to `ServerlessTransactionHandler` and `AWSTransactionWriter` to `ServerlessTransactionWriter`.\n- Adjusted error messages and comments to align with the new naming conventions.\n- Modified installation functions to use Serverless terminology for clarity and consistency across the codebase.",
          "timestamp": "2025-08-28T18:24:22+08:00",
          "tree_id": "93e7dea29a0988d46ac13dc06f749f60a12ae632",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/604da5d495eb42280077c9039dc1abbbad46a454"
        },
        "date": 1756376836560,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 136749625,
            "unit": "ns/op\t 1569755 B/op\t   27517 allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 136749625,
            "unit": "ns/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569755,
            "unit": "B/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.51,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "973396792 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.51,
            "unit": "ns/op",
            "extra": "973396792 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "973396792 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "973396792 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "anatolii@clearhavencm.com",
            "name": "toli-ch",
            "username": "toli-ch"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "880dc703453d2ea503bb38d8c1696cfea9dd26ba",
          "message": "Fix OS and architecture detection for Windows (#432)",
          "timestamp": "2025-09-01T13:21:29+08:00",
          "tree_id": "98a3583ca23297f52e25513875761d8a106008d7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/880dc703453d2ea503bb38d8c1696cfea9dd26ba"
        },
        "date": 1756704329199,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141628345,
            "unit": "ns/op\t 1569289 B/op\t   27517 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141628345,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569289,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.87,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "943166908 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.87,
            "unit": "ns/op",
            "extra": "943166908 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "943166908 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "943166908 times\n4 procs"
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
          "id": "5c57624f5dade94f066b36403f2b121177018c95",
          "message": "docs: add comprehensive development documentation (#433)",
          "timestamp": "2025-09-01T13:30:42+08:00",
          "tree_id": "7077202a5160074bc9496fc4306d482d83c9f61a",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/5c57624f5dade94f066b36403f2b121177018c95"
        },
        "date": 1756704883885,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137425763,
            "unit": "ns/op\t 1569572 B/op\t   27517 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137425763,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569572,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.85,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969734533 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.85,
            "unit": "ns/op",
            "extra": "969734533 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969734533 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969734533 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a4c32d32eaf936da574b86c2dff911b919e6c721",
          "message": "feat: propagate plugin unique identifier to serverless control plane (#428)\n\n* feat: propagate plugin unique identifier to serverless\n\n* resolve conflict",
          "timestamp": "2025-09-03T11:01:48+08:00",
          "tree_id": "11b239790c156b8130abde395ab7580f5f7d50d2",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/a4c32d32eaf936da574b86c2dff911b919e6c721"
        },
        "date": 1756868739208,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141548500,
            "unit": "ns/op\t 1569397 B/op\t   27517 allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141548500,
            "unit": "ns/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569397,
            "unit": "B/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "247 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.53,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "972546351 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.53,
            "unit": "ns/op",
            "extra": "972546351 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "972546351 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "972546351 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b4da677c5e88beeda4a93a6d2295b113dd2223c8",
          "message": "bump dify-cloud-kit version to v0.1.0 (#442)",
          "timestamp": "2025-09-03T18:08:28+08:00",
          "tree_id": "d53d3cb4dac25c4f13c873e6b35049a547f9fdac",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b4da677c5e88beeda4a93a6d2295b113dd2223c8"
        },
        "date": 1756894347718,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138368366,
            "unit": "ns/op\t 1569248 B/op\t   27516 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138368366,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569248,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.58,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981779443 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.58,
            "unit": "ns/op",
            "extra": "981779443 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981779443 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981779443 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "blackxin55+@gmail.com",
            "name": "kenwoodjw",
            "username": "kenwoodjw"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "bb83075acdafb60ee4f46451faca051ecf8387c8",
          "message": "feat: add max-size flag (#445)\n\nSigned-off-by: kenwoodjw <blackxin55+@gmail.com>",
          "timestamp": "2025-09-12T19:26:29+08:00",
          "tree_id": "0ab3206bc502a47f138fccbc97c9677708a0593e",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/bb83075acdafb60ee4f46451faca051ecf8387c8"
        },
        "date": 1757676635366,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 142895713,
            "unit": "ns/op\t 1569279 B/op\t   27517 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 142895713,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569279,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.92,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "978797227 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.92,
            "unit": "ns/op",
            "extra": "978797227 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "978797227 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "978797227 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e6bb4967a877e79a61da5043194b14b82860d63c",
          "message": "fix lock release (#447)",
          "timestamp": "2025-09-12T19:30:01+08:00",
          "tree_id": "ba2346226aa19372f42d1b42c3e211d829747d51",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/e6bb4967a877e79a61da5043194b14b82860d63c"
        },
        "date": 1757676840059,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 136158651,
            "unit": "ns/op\t 1569465 B/op\t   27517 allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 136158651,
            "unit": "ns/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569465,
            "unit": "B/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.16,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997241552 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.16,
            "unit": "ns/op",
            "extra": "997241552 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997241552 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997241552 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9e3081bd17ed8a17eb03e8bf913320972441f700",
          "message": "feat: support orphan plugin (#440)\n\n* support orphan plugin in serverless mode\n\n* add validation to UpgradePlugin\n\n* log reinstalls",
          "timestamp": "2025-09-12T20:34:38+08:00",
          "tree_id": "49ae790bc1a04ef59fa6bd9216d743e10960fe55",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9e3081bd17ed8a17eb03e8bf913320972441f700"
        },
        "date": 1757680719670,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138168789,
            "unit": "ns/op\t 1569095 B/op\t   27517 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138168789,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569095,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.56,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "971132616 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.56,
            "unit": "ns/op",
            "extra": "971132616 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "971132616 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "971132616 times\n4 procs"
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
          "id": "878edde4559a2955fdb785004d3a53ce91d02b96",
          "message": "feat(datasource): Implement datasource validation and invocation steps (#295)\n\n* feat(datasource): Implement datasource validation and invocation steps\n\n- Added new functionality for validating datasource credentials and invoking the first and second steps of the datasource process.\n- Introduced new API endpoints for datasource operations in the HTTP server.\n- Created corresponding service and controller methods to handle datasource requests.\n- Defined new request and response structures for datasource operations in the entities package.\n\n* feat: add routes\n\n* feat(datasource): Add initial datasource models and validation\n\n- Introduced new `DatasourceInstallation` model to represent datasource installations with relevant fields.\n- Created `datasource.go` file for future datasource service implementation.\n- Enhanced `datasource_declaration.go` with new types and validation functions for datasource provider and parameters.\n- Updated `plugin_declaration.go` to include datasource support in plugin structures.\n- Refactored `agent_declaration.go` and `tool_declaration.go` to use a unified `ParameterOption` type for options.\n\n* refactor(datasource): Update request and response types for online document content retrieval\n\n- Renamed and updated request and response types for the `DatasourceGetOnlineDocumentPageContent` function to improve clarity and consistency.\n- Introduced `RequestInvokeOnlineDocumentDatasourceGetContent` and `DatasourceInvokeOnlineDocumentGetContentResponse` types.\n- Adjusted related function signatures and dispatchers to reflect the new types across the datasource service and controller implementations.\n\n* feat(datasource): Implement datasource installation handling\n\n- Added functionality to create and update `DatasourceInstallation` records during plugin installation and upgrade processes.\n- Enhanced the `InstallPlugin` function to create a new datasource installation if a datasource declaration is present.\n- Updated the `UpgradePlugin` function to handle the deletion of the original datasource installation and creation of a new one if the datasource declaration changes.\n\n* feat(datasource): Add datasource registration handling in plugin runtime\n\n- Introduced handling for datasource declarations in the plugin runtime.\n- Updated `RemotePluginRuntime` to track datasource registration status.\n- Enhanced message processing to include validation and assignment of datasource declarations.\n\n* feat(datasource): Implement datasource listing and retrieval endpoints\n\n- Added `ListDatasources` and `GetDatasource` functions to the service layer for handling datasource queries.\n- Created corresponding controller methods to process HTTP requests for listing and retrieving datasources.\n- Implemented request validation for tenant ID, page, page size, plugin ID, and provider parameters.\n\n* add datasource routes\n\n* feat(datasource): Add icon remapping for datasource declarations\n\n- Implemented functionality to remap icons for both the main datasource and its individual datasources within the plugin declaration.\n- Enhanced error handling to provide clearer feedback when remapping fails.\n\n* fix(datasource): update OAuthSchema validation to remove unnecessary 'dive' tag\n\n- Modified the validation tag for OAuthSchema in DatasourceProviderDeclaration to simplify the validation process by removing the 'dive' requirement.\n\n* feat(tests): add datasource declaration parsing and output\n\n- Introduced a new main.go file for testing datasource declaration parsing.\n- Implemented JSON unmarshalling for RemotePluginRegisterPayload and DatasourceProviderDeclaration.\n- Added a sample datasource declaration for testing purposes.\n\n* fix(plugin): improve error handling in UninstallPlugin and add datasource deletion\n\n- Enhanced error handling to return a specific message when a plugin is not installed.\n- Added functionality to delete the associated datasource installation during the uninstallation process.\n\n* feat(datasource): add support for decoding datasource provider declaration\n\n- Added custom JSON marshalling and unmarshalling methods to handle CredentialsSchema and Datasources more effectively.\n- Improved error handling during YAML unmarshalling to support both object and array formats for CredentialsSchema.\n- Ensured proper initialization of DatasourceFiles and Tags to prevent nil references.\n\n* fix: provider type\n\n* refactor(datasource): simplify JSON unmarshalling for CredentialsSchema\n\n- Removed complex handling of CredentialsSchema in the UnmarshalJSON method, focusing on the Datasources field.\n- Streamlined the code to improve readability and maintainability by eliminating unnecessary checks and logic related to CredentialsSchema.\n\n* feat: streaming datasource\n\n* feat: datasource\n\n* feat: datasource\n\n* feat:datasource\n\n* feat:datasource\n\n* feat:datasource\n\n* feat: add redirect_uri field to OAuth request structs\n\n* feat: add online driver file request and response structures\n\n* feat: add online driver file request and response structures\n\n* feat: add online_driver datasource type to validation\n\n* feat: rename online driver to online drive and update related classes and methods :)\n\n* refactor: rename OnlineDocumentPageChunk to DatasourceGetPagesResponse and update related references\n\n* feat: update request types for online drive browsing and downloading\n\n* feat: add metadata field to OAuthGetCredentialsResult\n\n* feat: built-in json schema definations\n\n* feat(plugin_entities): add built-in schema definitions and processing for datasource YAML\n\n* test(plugin_entities): add unit tests for schema definitions and YAML processing\n\n* refactor(plugin_entities): centralize built-in schema definitions and processing\n\n* refactor(plugin_entities): remove unused properties from built-in schema definitions\n\n* refactor(plugin_entities): built-in schema & new datasource structure\n\n* refactor(plugin_entities): enhance schema processing with  checks and error handling\n\n* refactor(plugin_entities): update validation for OnlineDriveBrowseFilesRequest prefix field to be optional\n\n* refactor(json_schema): remove json schema definitions and validation\n\n* fix(plugin_entities): remove output_schema validation tests\n\n* feat: add rag tag\n\n---------\n\nCo-authored-by: Harry <xh001x@hotmail.com>\nCo-authored-by: Dongyu Li <544104925@qq.com>\nCo-authored-by: Novice <novice12185727@gmail.com>",
          "timestamp": "2025-09-16T17:32:14+08:00",
          "tree_id": "b2fefb814d803f68b6e32ecb697742f20015dc22",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/878edde4559a2955fdb785004d3a53ce91d02b96"
        },
        "date": 1758015378688,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138501775,
            "unit": "ns/op\t 1569414 B/op\t   27517 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138501775,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569414,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.86,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969006132 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.86,
            "unit": "ns/op",
            "extra": "969006132 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969006132 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969006132 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "jingfelix@outlook.com",
            "name": "Tianyi Jing",
            "username": "jingfelix"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1dacb201ff6e51f1b782f6e66263e57409542c33",
          "message": "feat(cmd): add datasource template (#449)\n\n* feat: add template for datasource plugins\n\nSigned-off-by: jingfelix <jingfelix@outlook.com>\n\n* fix: description and missing category\n\nSigned-off-by: jingfelix <jingfelix@outlook.com>\n\n---------\n\nSigned-off-by: jingfelix <jingfelix@outlook.com>",
          "timestamp": "2025-09-17T09:40:09+08:00",
          "tree_id": "5462c183cf6b4b837721f7500c798d7eabe977a3",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1dacb201ff6e51f1b782f6e66263e57409542c33"
        },
        "date": 1758073453178,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141731784,
            "unit": "ns/op\t 1569578 B/op\t   27517 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141731784,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569578,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.12,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "966010688 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.12,
            "unit": "ns/op",
            "extra": "966010688 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "966010688 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "966010688 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "100913391+crazywoola@users.noreply.github.com",
            "name": "crazywoola",
            "username": "crazywoola"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "dbbc4a29ccb13947d7bf2678aad1dc8c74d64a70",
          "message": "454 bump cli template (#455)\n\n* feat: update template\n\n* feat: update ja_JP\n\n* feat: update GUIDE.md",
          "timestamp": "2025-09-22T14:41:47+08:00",
          "tree_id": "70cca14091834b37934dcac567422316c8ed8788",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/dbbc4a29ccb13947d7bf2678aad1dc8c74d64a70"
        },
        "date": 1758523544669,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141444104,
            "unit": "ns/op\t 1568932 B/op\t   27517 allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141444104,
            "unit": "ns/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1568932,
            "unit": "B/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "246 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.55,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "953471497 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.55,
            "unit": "ns/op",
            "extra": "953471497 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "953471497 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "953471497 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "oleg.sadov@gmail.com",
            "name": "Oleg Sadov",
            "username": "sadov"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "294a521d54ca2adf69008854c56290e251c1f079",
          "message": "Added git to local.dockerfile (#453)\n\nSigned-off-by: Oleg Sadov <oleg.sadov@gmail.com>",
          "timestamp": "2025-09-23T16:14:24+08:00",
          "tree_id": "0e6ffe9b717a2a6cb69b1a116df25765aeba56ac",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/294a521d54ca2adf69008854c56290e251c1f079"
        },
        "date": 1758615509432,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 143366027,
            "unit": "ns/op\t 1569479 B/op\t   27517 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 143366027,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569479,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.84,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982948614 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.84,
            "unit": "ns/op",
            "extra": "982948614 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982948614 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982948614 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "64738319b8cbc0dd0eeeb2cba91f93a838810f73",
          "message": "Fix/json schema (#460)\n\n* fix(agent_declaration): remove unnecessary 'json_schema' validation from OutputSchema\n\n* fix(endpoint): add error handling for missing endpoint configuration in ListPluginEndpoints\n\n* fix(endpoint): update error handling for missing endpoint configuration in ListPluginEndpoints",
          "timestamp": "2025-09-23T23:53:21+08:00",
          "tree_id": "a8dce7972dd2b882c18403ac864dd3b0c5ac26d7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/64738319b8cbc0dd0eeeb2cba91f93a838810f73"
        },
        "date": 1758643034316,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139063073,
            "unit": "ns/op\t 1569736 B/op\t   27517 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139063073,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569736,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.57,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "990837170 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.57,
            "unit": "ns/op",
            "extra": "990837170 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "990837170 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "990837170 times\n4 procs"
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
          "id": "b85253f773a51e39418dd2b7f7685c63421caad2",
          "message": "feat(plugin): implement unauthorized Langgenius plugin blocking (#458)\n\n* feat(plugin): implement unauthorized Langgenius plugin blocking\n\n- Added configuration option to disable installation of plugins falsely claiming Langgenius authorship.\n- Introduced new error handling for unauthorized Langgenius claims during plugin installation.\n- Implemented tests to validate unauthorized Langgenius detection logic.\n- Updated environment configuration and service files to support the new feature.\n\n* fix: typo\n\n* refactor(plugin): update Langgenius plugin signature enforcement\n\n- Renamed configuration option from DISABLE_UNAUTHORIZED_LANGGENIUS_PACKAGE to ENFORCE_LANGGENIUS_PLUGIN_SIGNATURES for clarity.\n- Updated error messages and logic in the plugin installation process to reflect the new configuration.\n- Enhanced tests to validate the behavior of unauthorized Langgenius plugin detection with the new enforcement setting.",
          "timestamp": "2025-09-24T11:28:41+08:00",
          "tree_id": "041f34003dc4c1c83749a9a71e8880781a31d7d6",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b85253f773a51e39418dd2b7f7685c63421caad2"
        },
        "date": 1758684759772,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139310303,
            "unit": "ns/op\t 1569546 B/op\t   27516 allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139310303,
            "unit": "ns/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569546,
            "unit": "B/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "254 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.92,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "966922350 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.92,
            "unit": "ns/op",
            "extra": "966922350 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "966922350 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "966922350 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "obelisk.reg+git@gmail.com",
            "name": "QuantumGhost",
            "username": "QuantumGhost"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "50ff9a31ee865ba9d8f222b34277901e6c7dd449",
          "message": "packaging: add dependencies required by bowenliang123/md_exporter plugin (#466)\n\nsimplify local.dockerfile with heredoc",
          "timestamp": "2025-09-29T18:12:04+08:00",
          "tree_id": "15045ded4045ebd8ef4ce21adaf4384ece57d8db",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/50ff9a31ee865ba9d8f222b34277901e6c7dd449"
        },
        "date": 1759140969000,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138933243,
            "unit": "ns/op\t 1569945 B/op\t   27517 allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138933243,
            "unit": "ns/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569945,
            "unit": "B/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "249 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.8,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "967119381 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.8,
            "unit": "ns/op",
            "extra": "967119381 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "967119381 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "967119381 times\n4 procs"
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
          "id": "057d8ec0e423ebce267afc34b61995216005a98c",
          "message": "Check manifest when saving plugin packages (#467)",
          "timestamp": "2025-09-29T18:17:06+08:00",
          "tree_id": "ee14838e360fa1b56fc9927f34fc24c7461eb625",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/057d8ec0e423ebce267afc34b61995216005a98c"
        },
        "date": 1759141266178,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137415431,
            "unit": "ns/op\t 1569382 B/op\t   27517 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137415431,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569382,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.55,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "972570916 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.55,
            "unit": "ns/op",
            "extra": "972570916 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "972570916 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "972570916 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "40541269+CrabSAMA@users.noreply.github.com",
            "name": "CrabSAMA",
            "username": "CrabSAMA"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "2d81bb254ae704470c3edb6aebc612df6a58364c",
          "message": "fix: datasource plugin template typo (#464)",
          "timestamp": "2025-09-29T19:03:59+08:00",
          "tree_id": "07512409bafc1ff25b587baecff405a774d73a01",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/2d81bb254ae704470c3edb6aebc612df6a58364c"
        },
        "date": 1759144081740,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 141398507,
            "unit": "ns/op\t 1569659 B/op\t   27517 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 141398507,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569659,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.42,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "945518419 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.42,
            "unit": "ns/op",
            "extra": "945518419 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "945518419 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "945518419 times\n4 procs"
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
          "id": "96532cb1002db6493805054cf05d028363a52414",
          "message": "Remove Claude code review workflow (#473)",
          "timestamp": "2025-10-14T15:42:32+08:00",
          "tree_id": "1079cbdc25f179eaa3f35179b3cd9866e734e5d9",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/96532cb1002db6493805054cf05d028363a52414"
        },
        "date": 1760427995218,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 135196104,
            "unit": "ns/op\t 1569354 B/op\t   27517 allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 135196104,
            "unit": "ns/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569354,
            "unit": "B/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "253 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.31,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984775656 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.31,
            "unit": "ns/op",
            "extra": "984775656 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984775656 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984775656 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "511800046@qq.com",
            "name": "dancingMonkey",
            "username": "dancing-monkey"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9d7a48d3d1e9ba3b125e51d641ff8edd091093f2",
          "message": "Invalidate cache when enable endpoint. (#472)\n\nCo-authored-by: chengyi <chengyi02@corp.netease.com>",
          "timestamp": "2025-10-15T19:27:05+08:00",
          "tree_id": "795b2858b50f071bb76fc93b051b1f7f8afb8edc",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9d7a48d3d1e9ba3b125e51d641ff8edd091093f2"
        },
        "date": 1760527867732,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 140298827,
            "unit": "ns/op\t 1569811 B/op\t   27517 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 140298827,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569811,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.1,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "977839221 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.1,
            "unit": "ns/op",
            "extra": "977839221 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "977839221 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "977839221 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "123lzs123@gmail.com",
            "name": "Tonlo",
            "username": "TonloOO"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "b06732f2105b470a9e931b9d804c258f025873f6",
          "message": "fix(plugin): support customized stdio buffer size, align serverless runtime with local runtime (#470)\n\n* fix(plugin): support customized stdio buffer size, align serverless runtime with local runtime\n\n* fix(plugin): update config naming to pluginRuntimeBufferSize & pluginRuntimeMaxBufferSize, and keep the compatibility to stdoutBufferSize\n\n* fix: provide a default value for PLUGIN_RUNTIME_BUFFER_SIZE & PLUGIN_RUNTIME_MAX_BUFFER_SIZE\n\n* fix: remove redundant logic\n\n* fix: add compatibility to PLUGIN_STDIO_BUFFER_SIZE & PLUGIN_STDIO_MAX_BUFFER_SIZE",
          "timestamp": "2025-10-15T19:28:21+08:00",
          "tree_id": "0dfefe9c5866a253e302ee7ccdcfa5c1f21ce222",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/b06732f2105b470a9e931b9d804c258f025873f6"
        },
        "date": 1760527939101,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 136615963,
            "unit": "ns/op\t 1569124 B/op\t   27516 allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 136615963,
            "unit": "ns/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569124,
            "unit": "B/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.05,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "997007406 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.05,
            "unit": "ns/op",
            "extra": "997007406 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "997007406 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "997007406 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "byron@dify.ai",
            "name": "Byron.wang",
            "username": "41tair"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "a386efd44672dfcb7f538ea2c7cb666d89ec6a98",
          "message": "build serverless mode image with no-root permission (#463)",
          "timestamp": "2025-10-17T13:50:09+08:00",
          "tree_id": "6612c2a8ab51dd6e71c3ff88d172cac85e40f7ae",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/a386efd44672dfcb7f538ea2c7cb666d89ec6a98"
        },
        "date": 1760680446634,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137926327,
            "unit": "ns/op\t 1569584 B/op\t   27517 allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137926327,
            "unit": "ns/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569584,
            "unit": "B/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "258 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.01,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "964167957 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.01,
            "unit": "ns/op",
            "extra": "964167957 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "964167957 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "964167957 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "132866042+AlexChim1231@users.noreply.github.com",
            "name": "Alex Chim",
            "username": "AlexChim1231"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "1f910b867ccf93cbfc098eec958294e69ec64302",
          "message": "Fixes #479 Add filename for multi modal (#480)\n\n* Update llm.go\n\n* Update pkg/entities/model_entities/llm.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n* formatting\n\n* add unit test\n\n---------\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>",
          "timestamp": "2025-10-20T15:42:54+08:00",
          "tree_id": "f8ad2c55440ec1373b70319088a18aa1cf90286a",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/1f910b867ccf93cbfc098eec958294e69ec64302"
        },
        "date": 1760946413666,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 137616177,
            "unit": "ns/op\t 1569516 B/op\t   27517 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 137616177,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569516,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.9,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969851760 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.9,
            "unit": "ns/op",
            "extra": "969851760 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969851760 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969851760 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4589841b0c31ce199d5595d09730a835a539a1ee",
          "message": "feat: introduce trigger (#482)\n\n* feat(plugins): add FetchPluginReadme endpoint and update launch configurations\n\n* feat: add PluginReadme database model\n\n* feat: implement readme extracting and storage\n\n* feat: implement readme endpoint\n\n* feat: add plugin asset extraction endpoint with caching support\n\n* Implement trigger functionality and clean up dynamic select code\n\n- Added new trigger-related access types and actions in access.go.\n- Introduced new HTTP routes for trigger operations in http_server.gen.go.\n- Updated plugin declaration to include triggers in plugin_entities.\n- Removed unused dynamic select service and controller files.\n- Enhanced message handling in debugging_runtime to support trigger registration.\n\nThis update enhances the plugin system by integrating trigger capabilities while cleaning up legacy code.\n\n* Refactor trigger-related types and enhance dynamic select functionality\n\n- Updated TriggerProviderIdentity and TriggerProviderConfiguration to improve structure and validation.\n- Renamed TriggerConfiguration to TriggerDeclaration for clarity.\n- Added CredentialType to RequestDynamicParameterSelect for better request handling.\n- Enhanced PluginDecoderHelper to read and unmarshal trigger files correctly.\n\nThese changes streamline the trigger system and improve the overall request handling in the plugin architecture.\n\n* Add trigger functionality and enhance database integration\n\n- Introduced TriggerInstallation model for managing trigger installations in the database.\n- Updated autoMigrate function to include trigger installations in the migration process.\n- Added new HTTP routes for listing and retrieving triggers in the HTTP server.\n- Implemented ListTriggers and GetTrigger controller functions for handling trigger requests.\n- Enhanced plugin management functions to create, update, and delete trigger installations during plugin lifecycle events.\n\nThese changes integrate trigger capabilities into the system, improving the overall plugin functionality and management.\n\n* feat: add remapping for trigger icons in MediaBucket\n\n- Enhanced the RemapAssets function to include remapping of trigger identity icons and dark icons.\n- Added error handling for remapping failures to ensure robust asset management.\n\nThese changes improve the handling of trigger assets within the plugin system, ensuring icons are correctly remapped as needed.\n\n* feat: add Multiple field to TriggerParameter for enhanced configuration\n\n- Introduced a new Multiple field in the TriggerParameter struct to allow for multiple values in trigger configurations.\n- This addition improves the flexibility of trigger parameters within the plugin system.\n\nThese changes enhance the capability of trigger parameters, enabling more complex configurations.\n\n* feat: add Multiple field to ProviderConfig for enhanced configuration\n\n- Introduced a new Multiple field in the ProviderConfig struct to allow for multiple values in provider configurations.\n- This addition improves the flexibility of provider options within the plugin system.\n\nThese changes enhance the capability of provider configurations, enabling more complex setups.\n\n* fix(plugin): update validation error messages in ManifestValidate method\n\n- Enhanced error messages in the ManifestValidate function to include 'trigger' in the validation checks for plugin declarations.\n- Updated logic to ensure that all relevant fields are considered when validating the presence of mutually exclusive parameters.\n\n* feat(trigger): add CHECKBOX parameter type to plugin entities  and refactor the trigger provider strcuture\n\n- Introduced a new CHECKBOX parameter type in constant.go for plugin entities.\n- Updated tool_declaration.go and trigger_declaration.go to include TOOL_PARAMETER_TYPE_CHECKBOX and TRIGGER_PARAMETER_TYPE_CHECKBOX respectively.\n- Enhanced validation logic to accommodate the new CHECKBOX type in parameter checks.\n\n* fix(trigger): update SubscriptionSchema validation in TriggerProviderDeclaration\n\n- Changed SubscriptionSchema validation from 'omitempty' to 'required' in TriggerProviderDeclaration to ensure it is always provided.\n- Updated SubscriptionConstructor field to be a pointer to allow for optional inclusion in the trigger provider configuration.\n\n* fix(trigger): rename ParametersSchema to Parameters in SubscriptionConstructor\n\n- Updated the SubscriptionConstructor struct to rename the ParametersSchema field to Parameters for consistency.\n- Adjusted related JSON and YAML marshaling logic to reflect the new field name, ensuring proper handling of trigger parameters.\n\n* refactor(trigger): enhance YAML unmarshalling for SubscriptionConstructor and SubscriptionSchema\n\n- Introduced a new helper function to convert YAML nodes to ProviderConfig lists, improving the handling of subscription_schema and credentials_schema.\n- Updated the UnmarshalYAML method to utilize the new function, simplifying the logic for parsing different YAML formats.\n- Ensured proper initialization of SubscriptionConstructor fields to prevent nil pointer dereferences.\n\n* fix(trigger): update SubscriptionConstructor validation in TriggerProviderDeclaration\n\n- Changed the validation for SubscriptionConstructor in TriggerProviderDeclaration from 'omitempty,dive' to 'omitempty' to simplify the validation logic.\n- Ensured that the SubscriptionConstructor field remains optional while maintaining its intended functionality.\n\n* refactor(trigger): rename Trigger to Event in plugin entities and related structures\n\n- Updated the naming conventions in trigger_declaration.go to replace 'Trigger' with 'Event' for better clarity and consistency.\n- Adjusted related types, validation functions, and unmarshalling logic to reflect the new 'Event' terminology.\n- Ensured that all references to triggers in the codebase are updated to events, including in the SubscriptionConstructor and response structures.\n\n* refactor(trigger): rename TriggerInvoke to TriggerInvokeEvent and update related structures\n\n- Renamed TriggerInvoke function and associated request/response types to TriggerInvokeEvent for improved clarity.\n- Updated routing and controller methods to reflect the new naming convention.\n- Ensured all references to the trigger invoke functionality are consistent with the new event terminology.\n\n* refactor(trigger): remove Subscription struct from trigger_declaration.go and update TriggerDispatchEventRequest\n\n- Removed the Subscription struct from trigger_declaration.go to streamline the codebase.\n- Added Credentials field to TriggerDispatchEventRequest for enhanced functionality and clarity.\n- Ensured that the changes maintain consistency with the existing naming conventions and structures.\n\n* fix(trigger): improve nil checks for SubscriptionConstructor in TriggerProviderDeclaration\n\n- Added nil checks for SubscriptionConstructor before accessing its fields to prevent potential nil pointer dereferences.\n- Ensured that Parameters and CredentialsSchema are initialized only if SubscriptionConstructor is not nil, enhancing code robustness.\n\n* fix(plugin): add recovery mechanism in OnTraffic to handle panics\n\n- Introduced a deferred function in OnTraffic to recover from panics, logging the error and stack trace for better debugging.\n- This enhancement improves the stability of the DifyServer by preventing crashes due to unexpected runtime errors.\n\n* feat(trigger): add Subscription field to TriggerInvokeEventRequest\n\n- Introduced a new Subscription field in the TriggerInvokeEventRequest struct to accommodate subscription data.\n- Ensured the field is marked as required, enhancing the request's functionality and validation requirements.\n\n* refactor(event): simplify EventDescription structure in EventDeclaration\n\n- Removed the EventDescription struct and replaced it with a direct I18nObject field in EventDeclaration.\n- This change streamlines the event configuration by reducing complexity while maintaining required validation for the description.\n\n* feat(trigger): add UserID field to TriggerDispatchEventResponse\n\n- Introduced a new UserID field in the TriggerDispatchEventResponse struct to include user identification in the response.\n- The field is marked as optional, enhancing the response's flexibility while maintaining existing functionality.\n\n* feat: add payload to TriggerDispatchEventResponse\n\n* fix\n\n* feat(trigger): update TriggerDispatchEventResponse structure\n\n* fix: avoid path collusion\n\n* fix: missing )\n\n* fix: query param\n\n* fix: form param\n\n* fix: remove redundant dynamic parameter access type\n\n* fix: remove dynamic parameter access type from validation\n\n* Update internal/server/controllers/plugins.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n---------\n\nCo-authored-by: Stream <Stream_2@qq.com>\nCo-authored-by: Yeuoly <admin@srmxy.cn>\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>",
          "timestamp": "2025-10-29T13:56:58+08:00",
          "tree_id": "c31f507fe68fc44be7ca18e3d78820d02a0d98ff",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/4589841b0c31ce199d5595d09730a835a539a1ee"
        },
        "date": 1761717661780,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138971349,
            "unit": "ns/op\t 1569491 B/op\t   27517 allocs/op",
            "extra": "256 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138971349,
            "unit": "ns/op",
            "extra": "256 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569491,
            "unit": "B/op",
            "extra": "256 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "256 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.85,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "967477395 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.85,
            "unit": "ns/op",
            "extra": "967477395 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "967477395 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "967477395 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "06afbeca89f12dba42b07843aba5e6e9a49686be",
          "message": "feat(trigger): add Cancelled field to TriggerInvokeEventResponse for enhanced request handling (#496)",
          "timestamp": "2025-11-05T14:18:20+08:00",
          "tree_id": "759501c1908f68cc357105163b2103101f32d510",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/06afbeca89f12dba42b07843aba5e6e9a49686be"
        },
        "date": 1762323731083,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 136688639,
            "unit": "ns/op\t 1569238 B/op\t   27516 allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 136688639,
            "unit": "ns/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569238,
            "unit": "B/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "255 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.38,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "970970695 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.38,
            "unit": "ns/op",
            "extra": "970970695 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "970970695 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "970970695 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "hjlarry@163.com",
            "name": "非法操作",
            "username": "hjlarry"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "70ba852ecc941b2949e1c9f01772a8e8a89b8f39",
          "message": "feat: cli tool support create trigger plugins (#485)\n\n* feat: cli tool support create trigger plugins\n\n* fix(trigger): update placeholder comment in SubscriptionConstructor for webhook registration\n\n---------\n\nCo-authored-by: Harry <xh001x@hotmail.com>",
          "timestamp": "2025-11-05T15:45:58+08:00",
          "tree_id": "f19c367e22f375b0fa1fda40de178657830239ad",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/70ba852ecc941b2949e1c9f01772a8e8a89b8f39"
        },
        "date": 1762328998597,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 138357760,
            "unit": "ns/op\t 1569109 B/op\t   27516 allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 138357760,
            "unit": "ns/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569109,
            "unit": "B/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27516,
            "unit": "allocs/op",
            "extra": "248 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.58,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969538563 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.58,
            "unit": "ns/op",
            "extra": "969538563 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969538563 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969538563 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "20fc156f5c99908b6bda5e58c02a2b81c9c2c020",
          "message": "remove platform in go build (#505)",
          "timestamp": "2025-11-18T13:50:40+08:00",
          "tree_id": "0624e88bbbf5860032db92b818e3ca6aa5a2c343",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/20fc156f5c99908b6bda5e58c02a2b81c9c2c020"
        },
        "date": 1763445276723,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 139292708,
            "unit": "ns/op\t 1569295 B/op\t   27517 allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 139292708,
            "unit": "ns/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569295,
            "unit": "B/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "250 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 37.21,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "962360102 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.21,
            "unit": "ns/op",
            "extra": "962360102 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "962360102 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "962360102 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "610087273@qq.com",
            "name": "sHuXnHs",
            "username": "shuxnhs"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ad4a7c4d738d150ccee86fc20e071359e6f12887",
          "message": "tx cos support endpoint (#506)",
          "timestamp": "2025-11-18T14:14:00+08:00",
          "tree_id": "0ac6513995a656ec17362244894f5948a84edbb7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ad4a7c4d738d150ccee86fc20e071359e6f12887"
        },
        "date": 1763446676751,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkLocalOpenAILLMInvocation",
            "value": 134373757,
            "unit": "ns/op\t 1569599 B/op\t   27517 allocs/op",
            "extra": "260 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - ns/op",
            "value": 134373757,
            "unit": "ns/op",
            "extra": "260 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - B/op",
            "value": 1569599,
            "unit": "B/op",
            "extra": "260 times\n4 procs"
          },
          {
            "name": "BenchmarkLocalOpenAILLMInvocation - allocs/op",
            "value": 27517,
            "unit": "allocs/op",
            "extra": "260 times\n4 procs"
          },
          {
            "name": "BenchmarkStream",
            "value": 36.35,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "969859608 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.35,
            "unit": "ns/op",
            "extra": "969859608 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "969859608 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "969859608 times\n4 procs"
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
          "id": "888ad788bccc16ff244fef11ce593bfa6fe9d360",
          "message": "refactor: plugin lifecycle control panel (#499)\n\n* refactor: introduce local plugin control panel and cleanup environment setup process\n\n* fix: args\n\n* refactor: new local runtime\n\n* temp: stash work for refactor on RemotePluginServer\n\n* refactor: unify local runtime lifetime and sperate init environment process\n\n* chore: add missing files\n\n* stash\n\n* refactor: local plugin lifetime control\n\n* refactor: complete installation process of control panel\n\n* refactor: adapt service layer to new controlpanel\n\n* refactor: pluginManager.Install\n\n* fix: add routine wrap to InstallServerless, avoid blocking main thread\n\n* feat: reinstall serverless runtime\n\n* chore: add comments to Reinstall and update confusing naming\n\n* refactor: unify install plugin service\n\n* refactor: add labels to debugging runtime\n\n* refactor: add getters to plugin manager\n\n* refactor: split install service to decode/install_task/install service\n\n* ???\n\n* refactor: adapt controllers\n\n* refactor: session write\n\n* refactor: session runtime\n\n* Refine install task orchestration (#501)\n\n* refactor: installing task\n\n* refactor cluster management, decouple lifetime management and cluster\n\n* fix cli test command\n\n* fix: cleanup TODO comments and implement GracefulStop for instance\n\n* feat: add logger to control panel\n\n* fix: multiple nil references\n\n* refactor: better lifetime control\n\n* refactor: better cycle interval\n\n* fix(LocalPluginRuntime): prevent returning err when it's not  error\n\n* fix: avoid adding empty PipExtraArgs\n\n* fix: missing errors in Environment init\n\n* fix: add truncateMessage to avoid db explosion\n\n* cleanup: better lifecycle management\n\n* fix: init status at the beginning of installation\n\n* optimize: GracefulStop for pluginInstance\n\n* refactor: tests\n\n* refactor: centralize routine labels (#504)\n\n* cleanup: RoutineKey\n\n* fix: init routine pool\n\n* fix: correctly handle cluster register error\n\n* fix: memory leak\n\n* fix: add \\n to instance write\n\n* fix(installer.go): set success to true after succeed for defer func\n\n* refactor\n\n* fix: missing cwd in testutils\n\n* fix: scaleup default runtime nums to 1 when testing\n\n* fix: localruntime appconfig in testing module\n\n* Update internal/core/local_runtime/load_balancing.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n* fix: more efficiency implement in installer_local.go\n\n* fix: returns after failing in onDebuggingRuntimeDisconnected\n\n* fix: returns after failing in onDebuggingRuntimeDisconnected\n\n* fix: splits tests\n\n* refactor: naming\n\n* refactor: manifest.VersionX\n\n* fix: adapt SetDefault to tests\n\n* fix: enforce use constants in DBType\n\n* fix: generate\n\n* fix: linter\n\n* cleanup tests\n\n* refactor: change  package to\n\n* cleanup: useless codes\n\n* Update internal/cluster/plugin.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n* cleanup\n\n* refactor: decouple connection_key management from debugging_time\n\n* refactor: confused naming\n\n* feat: recycle resources to adapt to https://github.com/langgenius/dify-plugin-daemon/pull/500\n\n* refactor: confusing redirecting\n\n* fix: support get serverless runtime\n\n* fix: race condition in Launching\n\n* fix: avoid ManifestValidate in first step of debugging handshake\n\n* fix: adding ReleaseAllLocks to finalizers\n\n* wtf: what a beautiful code\n\n* refactor: rename Stream.Async to Stream.Process\n\n* fix: kill process if daed instance was detected\n\n* fix: correctly handle failures\n\n* fix: consistence of difference interfaces\n\n* fix: add stacktrace to panic\n\n* fix: only trigger once  event\n\n* fix: ensure plugin runtime was shutdown\n\n* feat: cleanup install tasks\n\n* fix: add scale logs\n\n---------\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>",
          "timestamp": "2025-11-18T17:28:02+08:00",
          "tree_id": "82a34f96a08bf1263a3776a134c24074665253b5",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/888ad788bccc16ff244fef11ce593bfa6fe9d360"
        },
        "date": 1763458168996,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 37.53,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "982962924 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.53,
            "unit": "ns/op",
            "extra": "982962924 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "982962924 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "982962924 times\n4 procs"
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
          "id": "20a40526f24609daaacb1e82d4db17778f326a58",
          "message": "bump golang to 1.25 (#507)\n\n* bump golang to 1.25\n\n* bump",
          "timestamp": "2025-11-18T18:20:29+08:00",
          "tree_id": "2c1b23763214b8913dfa3debe71c159cd05182a7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/20a40526f24609daaacb1e82d4db17778f326a58"
        },
        "date": 1763461321521,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 37.43,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "946707218 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.43,
            "unit": "ns/op",
            "extra": "946707218 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "946707218 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "946707218 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "c1ff2515aee59a8ec8908fd7d4c5710c59fa52d8",
          "message": "fix: remove restriction of orphan plugins in local mode (#509)",
          "timestamp": "2025-11-24T17:34:32+08:00",
          "tree_id": "aa71bb226eda648fa50f01890cd01737f02cc77d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/c1ff2515aee59a8ec8908fd7d4c5710c59fa52d8"
        },
        "date": 1763977012659,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.5,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "988442043 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.5,
            "unit": "ns/op",
            "extra": "988442043 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "988442043 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "988442043 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "hjlarry@163.com",
            "name": "非法操作",
            "username": "hjlarry"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "e10b48fccd804fe7fd1522fe1cbc9060ea502080",
          "message": "fix: remote debug not work (#515)",
          "timestamp": "2025-12-02T16:12:17+08:00",
          "tree_id": "268ef38113b21a0b90f0c4380a48a6980c554352",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/e10b48fccd804fe7fd1522fe1cbc9060ea502080"
        },
        "date": 1764663278341,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 39.68,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "906754461 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 39.68,
            "unit": "ns/op",
            "extra": "906754461 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "906754461 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "906754461 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "hjlarry@163.com",
            "name": "非法操作",
            "username": "hjlarry"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "4b25ad67bac565eb1f42d31cbf9e2d59a5fe1038",
          "message": "fix: tool and trigger not response icon_dark (#517)",
          "timestamp": "2025-12-02T16:12:51+08:00",
          "tree_id": "4e13fd9a0b583706c79c60c77b8fc75dbadb2373",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/4b25ad67bac565eb1f42d31cbf9e2d59a5fe1038"
        },
        "date": 1764663318681,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.54,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "990280719 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.54,
            "unit": "ns/op",
            "extra": "990280719 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "990280719 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "990280719 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "BorisPolonsky@users.noreply.github.com",
            "name": "Boris Polonsky",
            "username": "BorisPolonsky"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "d3d1a652e65f3eff006368b513207aefb1594b4d",
          "message": "Rename InstallToLocalFromPkg to InstallToLocal (#511)",
          "timestamp": "2025-12-02T20:19:44+08:00",
          "tree_id": "56970103f62abbb90f1fbc0a17a1b5a182867597",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/d3d1a652e65f3eff006368b513207aefb1594b4d"
        },
        "date": 1764678132580,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 37.01,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981622772 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.01,
            "unit": "ns/op",
            "extra": "981622772 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981622772 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981622772 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "f9546889309ee2c90dc01c5722712918bd4da466",
          "message": "enterprise related adjustment (#510)\n\n* remove enterprise logics\n\n* release lock if runtime already installed\n\n* support setting serverless endpoint by api\n\n* remove global tenant id totally\n\n* scan timeout tasks\n\n* adding comments\n\n* use index in loop\n\n* add log",
          "timestamp": "2025-12-03T18:13:54+08:00",
          "tree_id": "f4108d5817e4513e6252f25a0d13aac4056d7d3f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/f9546889309ee2c90dc01c5722712918bd4da466"
        },
        "date": 1764756986140,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.42,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984699484 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.42,
            "unit": "ns/op",
            "extra": "984699484 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984699484 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984699484 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "fatelei@gmail.com",
            "name": "wangxiaolei",
            "username": "fatelei"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "68db8475c9a84014f23459708c5ab8d6238a4769",
          "message": "fix: fix ReadCloser is not closed (#518)",
          "timestamp": "2025-12-03T18:18:30+08:00",
          "tree_id": "e24b32dac50e2a9a0bb609f7c23ad8c55decb76e",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/68db8475c9a84014f23459708c5ab8d6238a4769"
        },
        "date": 1764757270139,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.36,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "993393021 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.36,
            "unit": "ns/op",
            "extra": "993393021 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "993393021 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "993393021 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "wylswz@163.com",
            "name": "Yunlu Wen",
            "username": "wylswz"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "77637afbc5ec0a15c9616850c891a038ba11eca6",
          "message": "fix: handle connection close on the connector side (#522)",
          "timestamp": "2025-12-07T18:02:44+08:00",
          "tree_id": "06e9f01a7f9bc0f3ca8802bf4ff929bff9e63cda",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/77637afbc5ec0a15c9616850c891a038ba11eca6"
        },
        "date": 1765101906168,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 37.63,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "964405075 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.63,
            "unit": "ns/op",
            "extra": "964405075 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "964405075 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "964405075 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "hjlarry@163.com",
            "name": "非法操作",
            "username": "hjlarry"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "89ebb771887b8d02178d849ac3cc3a247f55b24c",
          "message": "feat: to support tool multi-select input (#523)",
          "timestamp": "2025-12-08T12:32:57+08:00",
          "tree_id": "33401d7e929e8bd3df03a82b1f0df520285c9d09",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/89ebb771887b8d02178d849ac3cc3a247f55b24c"
        },
        "date": 1765168524597,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 39.43,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "906485782 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 39.43,
            "unit": "ns/op",
            "extra": "906485782 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "906485782 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "906485782 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "longbing.ljw@oceanbase.com",
            "name": "longbingljw",
            "username": "longbingljw"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "ad7d3a1ab7be32c59b7b57d8bf7118b07ed98c56",
          "message": "update config after mysql adaptation (#514)\n\n* update config after mysql adaptation\n\n* fix\n\n* fix",
          "timestamp": "2025-12-08T16:38:26+08:00",
          "tree_id": "cd403b12452400031580438be55ec608743febcd",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/ad7d3a1ab7be32c59b7b57d8bf7118b07ed98c56"
        },
        "date": 1765183259933,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.65,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "987841123 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.65,
            "unit": "ns/op",
            "extra": "987841123 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "987841123 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "987841123 times\n4 procs"
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
          "id": "0d282de6a1740335370956a267e6a20e8ae09227",
          "message": "feat: support multimodal rerank and embedding (#503)\n\n* fix: align multimodal permission mapping\n\n* fix: implement mock interface for multimodal embeddings\n\n* fix: support multimodal embedding\n\n* fix: incorrect reference",
          "timestamp": "2025-12-09T14:56:11+08:00",
          "tree_id": "669c78f09c9a4961fae217c23cf2131d1c88447e",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/0d282de6a1740335370956a267e6a20e8ae09227"
        },
        "date": 1765263515491,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 37.2,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "970659117 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 37.2,
            "unit": "ns/op",
            "extra": "970659117 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "970659117 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "970659117 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8e5743aef671e8736b33b31a18658e246803204c",
          "message": "feat: implement locking mechanism for concurrent plugin installation and upgrade (#526)\n\n* add locking to prevent simultaneous installations of the same plugin\n* ensure proper unlocking of keys in case of errors during installation and upgrade\n* handle database not found error in DeletePluginInstallationItemFromTask",
          "timestamp": "2025-12-10T17:27:43+08:00",
          "tree_id": "1851307b9b63f90ce439dd6958cbb7f5cb261932",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/8e5743aef671e8736b33b31a18658e246803204c"
        },
        "date": 1765359018954,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.46,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984624910 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.46,
            "unit": "ns/op",
            "extra": "984624910 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984624910 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984624910 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "xh001x@hotmail.com",
            "name": "Maries",
            "username": "Mairuis"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "96b51115cb30f008bf4eda7e3787ea27d39c18e2",
          "message": "fix: add unique index for plugin installation to prevent duplicate entries (#528)",
          "timestamp": "2025-12-10T18:54:25+08:00",
          "tree_id": "774e326f33ad0a66a226bd71ed0d972967069a1d",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/96b51115cb30f008bf4eda7e3787ea27d39c18e2"
        },
        "date": 1765364208384,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.29,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "984446544 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.29,
            "unit": "ns/op",
            "extra": "984446544 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "984446544 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "984446544 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "49699333+dependabot[bot]@users.noreply.github.com",
            "name": "dependabot[bot]",
            "username": "dependabot[bot]"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "9cdcdb492e766ab20b36f27e8cf10c2891d12a53",
          "message": "chore(deps): bump golang.org/x/crypto from 0.40.0 to 0.45.0 (#531)",
          "timestamp": "2025-12-12T12:06:57+08:00",
          "tree_id": "a69f6446edd432eedaf9e246f2fb7c87543aea52",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/9cdcdb492e766ab20b36f27e8cf10c2891d12a53"
        },
        "date": 1765512557816,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 31.01,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 31.01,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "0eb2cbe2847577a75d8d411071724355202afdaf",
          "message": "Fix/remove tenant id reference in readme service (#530)\n\n* fix: remove tenant ID reference from plugin readme service\n\n* fix: rename PluginReadme to PluginReadmeRecord to remove tenant ID reference\n\n* fix: handle error when saving plugin readme map to database\n\n* fix: remove TenantId from FetchPluginReadme request binding\n\n* fix: standardize error message casing for plugin unique identifier validation",
          "timestamp": "2025-12-12T15:35:06+08:00",
          "tree_id": "b7b4257ad570a51c7782b9ff16814d504c364079",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/0eb2cbe2847577a75d8d411071724355202afdaf"
        },
        "date": 1765524983997,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 30.75,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 30.75,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "11e00ea0268a233a299e55ad4f827c5d8f87278e",
          "message": "refactor: remove uuid-ossp dependency and update ID generation logic (#536)\n\n* refactor: remove uuid-ossp dependency and update ID generation logic\n\n* fix: condition error",
          "timestamp": "2025-12-13T16:25:15+08:00",
          "tree_id": "bd3781df4ed56b2cc6609bb14c43b57d8828d70f",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/11e00ea0268a233a299e55ad4f827c5d8f87278e"
        },
        "date": 1765614396106,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 36.4,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "981869842 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 36.4,
            "unit": "ns/op",
            "extra": "981869842 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "981869842 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "981869842 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "hjlarry@163.com",
            "name": "非法操作",
            "username": "hjlarry"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "72e9a7bc21acd14f7994f0b496da6454b99ee7fc",
          "message": "fix: plugin log not display after refactor (#537)\n\n* fix: plugin log not display after refactor\n\n* Update internal/core/local_runtime/notifier_logger.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n* feat: add logging support for local runtime instance events\n\n---------\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\nCo-authored-by: Yeuoly <admin@srmxy.cn>",
          "timestamp": "2025-12-16T14:14:19+08:00",
          "tree_id": "0e15981e67ea9ef4607343e4725d27d2db73e7c4",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/72e9a7bc21acd14f7994f0b496da6454b99ee7fc"
        },
        "date": 1765865741998,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 31.17,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 31.17,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
          }
        ]
      },
      {
        "commit": {
          "author": {
            "email": "myg133@qq.com",
            "name": "Michael.Y.Ma",
            "username": "myg133"
          },
          "committer": {
            "email": "noreply@github.com",
            "name": "GitHub",
            "username": "web-flow"
          },
          "distinct": true,
          "id": "8ff1c2c9f203387d3e432372b137ad2d825720d1",
          "message": "Feat: Add OSS-specific parameters for HW and ALI private deployment (#539)\n\n* Update internal/types/app/config.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n* Update internal/server/server.go\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>\n\n---------\n\nCo-authored-by: gemini-code-assist[bot] <176961590+gemini-code-assist[bot]@users.noreply.github.com>",
          "timestamp": "2025-12-17T16:36:56+08:00",
          "tree_id": "09a79d0e07e8024140d3e92b62288cebf882f1a7",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/8ff1c2c9f203387d3e432372b137ad2d825720d1"
        },
        "date": 1765960757196,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 30.44,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 30.44,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "4b75eed6b7c14d1d755c3e9cac59c4d847bec2c1",
          "message": "refactor: separate CI tests, and support pg bouncer (#529)\n\n* feat: add support for PgBouncer in database initialization\n\n* refactor: consolidate gorm configuration for database connection\n\n* Add MySQL and multi-driver DB integration tests (#535)\n\n* feat: streamline integration tests by using a centralized docker-compose file\n\n* fix: add timeout\n\n* refactor: replace inline service definitions with docker-compose action for integration tests\n\n* fix: update pgbouncer image and environment variable names for consistency",
          "timestamp": "2025-12-17T22:08:47+08:00",
          "tree_id": "a83a987a0bd30204c67985e69ef98255d1ed2dcc",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/4b75eed6b7c14d1d755c3e9cac59c4d847bec2c1"
        },
        "date": 1765980669036,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 30.24,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 30.24,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
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
          "id": "38ec23dabf086bfcc4420a4174a6d67daf3ad794",
          "message": "fix: add default database for pg bouncer support (#545)",
          "timestamp": "2025-12-19T15:29:44+08:00",
          "tree_id": "395bf4ef3c2cef0f37e1c760039c364eb24deee5",
          "url": "https://github.com/langgenius/dify-plugin-daemon/commit/38ec23dabf086bfcc4420a4174a6d67daf3ad794"
        },
        "date": 1766129530270,
        "tool": "go",
        "benches": [
          {
            "name": "BenchmarkStream",
            "value": 30.59,
            "unit": "ns/op\t      15 B/op\t       0 allocs/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - ns/op",
            "value": 30.59,
            "unit": "ns/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - B/op",
            "value": 15,
            "unit": "B/op",
            "extra": "1000000000 times\n4 procs"
          },
          {
            "name": "BenchmarkStream - allocs/op",
            "value": 0,
            "unit": "allocs/op",
            "extra": "1000000000 times\n4 procs"
          }
        ]
      }
    ]
  }
}