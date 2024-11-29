class Dify < Formula
  desc "Dify Plugin Command Line Tool"
  homepage "https://github.com/langgenius/dify-plugin-daemon"
  version "0.1.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.intel?
    url "file://#{__dir__}/bin/dify-plugin-darwin-amd64.tar.gz"
    sha256 "3b0172bfdaf19396a855974b6f83e03a86ce2a073615cd7d6fbbb104c3d96946"
  elsif OS.mac? && Hardware::CPU.arm?
    url "file://#{__dir__}/bin/dify-plugin-darwin-arm64.tar.gz"
    sha256 "8a527f7bc61046aa11992d76cc2e3fe2a2c38cf3434d882273fcba30dd3a2e00"
  end

  def install
    if Hardware::CPU.intel?
      bin.install "dify-plugin-darwin-amd64" => "dify"
    else
      bin.install "dify-plugin-darwin-arm64" => "dify"
    end
  end

  test do
    system "#{bin}/dify", "--version"
  end
end