# Site Portal Frontend UI

This frontend UI of Site Portal is based on Clarity and Angular.

# Start a Dev Environment: 
1. Run `npm install` under "frontend" directory.
2. create "proxy.config.json" file under "frontend" directory, and replace the URL of target with your available backend server.
```
 {
    "/api/v1": {
      "target": "http://localhost:8080",
      "secure": false,
      "changeOrigin": true,
      "logLevel": "debug",
        "headers": {
            "Connection": "keep-alive"
        }
    }
  }
```
3. Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The app will automatically reload if you change any of the source files.
4. Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.