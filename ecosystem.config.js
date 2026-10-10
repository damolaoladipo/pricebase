module.exports = {
  apps: [
    {
      name: "pricebase-api-dev",
      cwd: __dirname,
      script: "go",
      args: "run ./cmd/pricebase",
      interpreter: "none",
      watch: true,

      ignore_watch: [".git", "bin", "tmp", "logs", "views", "public",  "internal/_data/logs",],
      
      restart_delay: 1000,
      env: {
        GO_ENV: "development",
      },
    },

    {
      name: "pricebase-api",
      cwd: __dirname,
      script: "./bin/pricebase-api",
      interpreter: "none",
      
      watch: false,
      env: {
        GO_ENV: "production",
      },
    },
  ],
};
