{ pkgs, lib, config, ... }:

{
  # Core packages
  packages = with pkgs; [
    git
    gomod2nix
    golangci-lint
    gotestsum
    air
  ];

  # Enable Go with delve support
  languages.go = {
    enable = true;
    enableHardeningWorkaround = true;  # Required for delve debugger
  };

  # Development server
  processes.app = {
    exec = "air";
  };

  # Common tasks
  tasks = {
    "dev:setup" = {
      exec = ''
        go mod download
        gomod2nix generate
      '';
    };

    "app:lint" = {
      exec = "golangci-lint run ./...";
    };

    "app:test" = {
      exec = "gotestsum --format=short-verbose -- -v -race $(go list ./... | grep -v '/docs/')";
    };

    "app:build" = {
      exec = "go build -o bin/shine ./cmd/shine";
    };
  };

  # Test automation
  enterTest = ''
    golangci-lint run ./...
    go test -v -race $(go list ./... | grep -v '/docs/')
  '';

  # Environment variables
  dotenv.enable = true;

  # Shell welcome message
  enterShell = ''
    mkdir -p bin
    export PATH="$DEVENV_ROOT/bin:$PATH"

    echo "🚀 Go $(go version | cut -d' ' -f3) development environment"
    echo ""
    echo "Available commands:"
    echo "  devenv up          - Start development server"
    echo "  devenv tasks run   - Run tasks"
    echo "  devenv test        - Run tests"
    echo ""
  '';

  # Git hooks for code quality
  git-hooks.hooks = {
    govet = {
      enable = true;
      pass_filenames = false;
    };
    gotest.enable = true;
    golangci-lint = {
      enable = true;
      pass_filenames = false;
    };
  };

}
