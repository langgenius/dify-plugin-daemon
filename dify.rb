class Dify < Formula
  desc "Dify"
  homepage "https://github.com/langgenius/dify-plugin-daemon"
  VERSION = "0.0.1"
  version VERSION

  if OS.mac?
    if Hardware::CPU.intel?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/#{VERSION}/dify-plugin-darwin-amd64"
    elsif Hardware::CPU.arm?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/#{VERSION}/dify-plugin-darwin-arm64"
    end
  elsif OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/#{VERSION}/dify-plugin-linux-amd64"
    elsif Hardware::CPU.arm?
      url "https://github.com/langgenius/dify-plugin-daemon/releases/download/#{VERSION}/dify-plugin-linux-arm64"
    end
  elsif OS.windows?
    url "https://github.com/langgenius/dify-plugin-daemon/releases/download/#{VERSION}/dify-plugin-windows-amd64"
  end

  def install
    # Determine the OS and architecture to select the correct binary.
    os = if OS.mac?
           "darwin"
         elsif OS.linux?
           "linux"
         elsif OS.windows?
           "windows"
         end

    arch = if Hardware::CPU.intel?
             "amd64"
           elsif Hardware::CPU.arm?
             "arm64"
           end

    bin.install "dify-plugin-#{os}-#{arch}" => "dify"
  end

  test do
    # Verify that running `dify --version` returns the expected version.
    assert_match VERSION.to_s, shell_output("#{bin}/dify --version")
  end
end
