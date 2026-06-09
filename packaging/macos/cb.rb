class Cb < Formula
  desc "Cross-device clipboard for developers"
  homepage "https://github.com/Morolis/cb"
  version "0.1.0"
  license "Apache-2.0"

  on_macos do
    on_intel do
      url "https://github.com/Morolis/cb/releases/download/v#{version}/cb-darwin-amd64.tar.gz"
      sha256 "UPDATE_AFTER_RELEASE"
    end

    on_arm do
      url "https://github.com/Morolis/cb/releases/download/v#{version}/cb-darwin-arm64.tar.gz"
      sha256 "UPDATE_AFTER_RELEASE"
    end
  end

  def install
    bin.install "cb-darwin-#{Hardware::CPU.arch}" => "cb"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/cb --version")
  end
end
