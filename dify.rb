class Dify < Formula
  desc "Dify Plugin Command Line Tool"
  homepage "https://github.com/langgenius/dify-plugin-daemon"
  version "0.1.0"
  license "MIT"

  if OS.mac? && Hardware::CPU.intel?
    url "file:///Users/minibanana/Program/projects/dify-plugin-daemon/bin/dify-plugin-darwin-amd64.tar.gz"
    sha256 "e01b770809c9f195524578cd88ca121c7f352a9eaf8187b07fe9596d9c3345ef"
  elsif OS.mac? && Hardware::CPU.arm?
    url "file:///Users/minibanana/Program/projects/dify-plugin-daemon/bin/dify-plugin-darwin-arm64.tar.gz"
    sha256 "8696eaebff598a49577e22ba893039fc3cdecaecf30954a98379b09c71ce4f9d"
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