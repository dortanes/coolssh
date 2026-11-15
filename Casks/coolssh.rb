cask "coolssh" do
  version "0.1.0"
  sha256 "e98f0a5af3aa47d4a02d2e44b7098d69dc3039ff12361d4c5ca7da8fdabe5659"

  url "https://github.com/dortanes/coolssh/releases/download/v#{version}/coolssh-#{version}-macos-universal.dmg"
  name "CoolSSH"
  desc "SSH client with host folders that pass settings down"
  homepage "https://github.com/dortanes/coolssh"

  livecheck do
    url :url
    strategy :github_latest
  end

  depends_on macos: :ventura

  app "CoolSSH.app"
end
