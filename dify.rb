class Dify < Formula
  desc "Dify Plugin Command Line Tool"
  homepage "https://github.com/langgenius/dify-plugin-daemon"
  version "0.1.0"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.intel?
      url "file://#{__dir__}/bin/dify-plugin-darwin-amd64.tar.gz"
      sha256 "e57c3b3adab56e43f588b5add4e95d4449d916f7ef2e09d6ff4ab760e22e7bd8"
    else
      url "file://#{__dir__}/bin/dify-plugin-darwin-arm64.tar.gz"
      sha256 "5a60e8a6faa43dc3241ca74856a95710d695f164d5e845bb71471b8db7ce50e7"
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
