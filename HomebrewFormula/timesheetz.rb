class Timesheetz < Formula
  desc "Timesheet management with TUI and REST API - write hours like a unicorn"
  homepage "https://github.com/joelgrimberg/timesheetz"
  version "1.5.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/joelgrimberg/timesheetz/releases/download/v1.5.0/timesheetz_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    else
      url "https://github.com/joelgrimberg/timesheetz/releases/download/v1.5.0/timesheetz_Darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_AMD64"
    end
  end

  def install
    bin.install "timesheet" => "timesheetz"
  end

  def post_install
    # Create config directory
    config_dir = "#{ENV['HOME']}/.config/timesheetz"
    system "mkdir", "-p", config_dir

    # Create launch agents directory
    launch_agents_dir = "#{ENV['HOME']}/Library/LaunchAgents"
    system "mkdir", "-p", launch_agents_dir

    # Create the launch agent plist
    plist_path = "#{launch_agents_dir}/com.timesheetz.plist"

    plist_content = <<~EOS
      <?xml version="1.0" encoding="UTF-8"?>
      <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
      <plist version="1.0">
      <dict>
          <key>Label</key>
          <string>com.timesheetz</string>
          <key>ProgramArguments</key>
          <array>
              <string>#{bin}/timesheetz</string>
              <string>--no-tui</string>
          </array>
          <key>RunAtLoad</key>
          <true/>
          <key>KeepAlive</key>
          <true/>
          <key>StandardOutPath</key>
          <string>#{ENV['HOME']}/Library/Logs/timesheetz.out</string>
          <key>StandardErrorPath</key>
          <string>#{ENV['HOME']}/Library/Logs/timesheetz.err</string>
          <key>EnvironmentVariables</key>
          <dict>
              <key>PATH</key>
              <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
              <key>HOME</key>
              <string>#{ENV['HOME']}</string>
          </dict>
      </dict>
      </plist>
    EOS

    File.write(plist_path, plist_content)
    system "chmod", "644", plist_path

    # Create logs directory
    system "mkdir", "-p", "#{ENV['HOME']}/Library/Logs"

    ohai "Timesheetz installed successfully!"
    ohai "To start the Launch Agent now, run:"
    ohai "  launchctl load ~/Library/LaunchAgents/com.timesheetz.plist"
    ohai ""
    ohai "The Launch Agent will start automatically on next login."
    ohai "To prevent auto-start, run:"
    ohai "  launchctl unload ~/Library/LaunchAgents/com.timesheetz.plist"
  end

  def caveats
    <<~EOS
      Timesheetz has been installed!

      The Launch Agent has been created at:
        ~/Library/LaunchAgents/com.timesheetz.plist

      To start Timesheetz now:
        launchctl load ~/Library/LaunchAgents/com.timesheetz.plist

      To check if it's running:
        launchctl list | grep timesheetz

      View logs at:
        ~/Library/Logs/timesheetz.out
        ~/Library/Logs/timesheetz.err

      Run the TUI from command line:
        timesheetz

      Run in development mode:
        timesheetz --dev

      For more options:
        timesheetz --help
    EOS
  end

  test do
    system "#{bin}/timesheetz", "--version"
  end
end
