class Dify < Formula
  desc "Dify Plugin Command Line Tool"
  homepage "https://github.com/langgenius/dify-plugin-daemon"
  version "0.1.0"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.intel?
      url "file://#{__dir__}/bin/dify-plugin-darwin-amd64.tar.gz"
      sha256 "39ab1029634acf1caa8e68efcc393162a0fc760170472071b5fc02d06b084993"
    else
      url "file://#{__dir__}/bin/dify-plugin-darwin-arm64.tar.gz"
      sha256 "467cd4d13a7be4d1583589da4cf6b39e3b72b2fe4a02b6bc59c2c36309459a4b"
    end
  else
    odie "This formula only supports macOS."
  end

  def install
    bin.install "dify-plugin-darwin-#{Hardware::CPU.arch}" => "dify"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dify --version")
  end
end
