window.BENCHMARK_DATA = {
  "lastUpdate": 1745310731780,
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
      }
    ]
  }
}