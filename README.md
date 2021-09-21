# dronesec
Simple wrapper to drone api that creates a new drone secret. Meant to be used in drone itself. Usage with enviromental variables, should use the preset drone variables for everything else except access token, which should be put behind "DRONE_TOKEN" in the cicd.

The container image will attempt to push all files with filename form: "drone_secret.secret" where the final extension is removed to create secret name and content is used as secret body.
