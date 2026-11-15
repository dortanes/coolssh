cask "coolssh" do
  version "0.0.0"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"

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
