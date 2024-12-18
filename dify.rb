class Dify < Formula
  desc "Dify"
  homepage "https://github.com/langgenius/dify-plugin-daemon"
  version "0.0.1-beta.20"

  if OS.mac?
    if Hardware::CPU.intel?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/0.0.1-beta.20/dify-plugin-darwin-amd64"
    elsif Hardware::CPU.arm?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/0.0.1-beta.20/dify-plugin-darwin-arm64"
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/0.0.1-beta.20/dify-plugin-linux-amd64"
    elsif Hardware::CPU.arm?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/0.0.1-beta.20/dify-plugin-linux-arm64"
    end
  elsif OS.windows?
    url "https://github.com/langgenius/dify-plugin-daemon/releases/download/0.0.1-beta.20/dify-plugin-windows-amd64"
  end

  def install
    bin.install "dify-plugin-darwin-#{Hardware::CPU.arch}" => "dify"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dify --version")
  end
end
