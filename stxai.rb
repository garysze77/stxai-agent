# Homebrew Formula for STX AI Agent
#
# To use:
#   brew install --formula stxai.rb
#
# Or via tap (after creating garysze77/homebrew-stxai):
#   brew tap garysze77/stxai-agent
#   brew install stxai
#
# Auto-published by GoReleaser once garysze77/homebrew-stxai repo exists.
# Uncomment the brews section in ../.goreleaser.yml to enable.

class Stxai < Formula
  desc "STX AI Agent — Autonomous Financial AI for US & HK stocks"
  homepage "https://github.com/garysze77/stxai-agent"
  version "0.10.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/garysze77/stxai-agent/releases/download/v#{version}/stxai_darwin_arm64.tar.gz"
      sha256 "" # updated by GoReleaser
    elsif Hardware::CPU.intel?
      url "https://github.com/garysze77/stxai-agent/releases/download/v#{version}/stxai_darwin_amd64.tar.gz"
      sha256 "" # updated by GoReleaser
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/garysze77/stxai-agent/releases/download/v#{version}/stxai_linux_arm64.tar.gz"
      sha256 "" # updated by GoReleaser
    elsif Hardware::CPU.intel?
      url "https://github.com/garysze77/stxai-agent/releases/download/v#{version}/stxai_linux_amd64.tar.gz"
      sha256 "" # updated by GoReleaser
    end
  end

  def install
    bin.install "stxai"
  end

  test do
    system "#{bin}/stxai", "--version"
  end
end
