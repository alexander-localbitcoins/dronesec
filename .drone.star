def main(ctx):
    return [
        {
            "kind": "pipeline",
            "type": "docker",
            "name": "dronesecret",
            "trigger": {"event": ["push"]},
            "steps": [
                test_step,
                build_step() if ctx.build.branch == "main" and ctx.build.event == "push" else build_step(True),
            ],
            "node": {"docker": "slow"},
        },
    ]

test_step = {
    "name": "tests",
    "image": "golang:latest",
    "environment": {
        "DRONE_TOKEN": {"from_secret": "drone_token"},
    },
    "commands": [
        "go test -v",
        "cd ./client/",
        "go test -v",
    ],
}

def build_step(test=False):
    step = {
        "name": "build-container",
        "image": "plugins/docker",
        "settings": {
            "username": {"from_secret": "github_username"},
            "password": {"from_secret": "github_secret"},
            "context": "./docker/",
            "dockerfile": "./docker/Dockerfile",
            "repo": "ghcr.io/localbitcoins/dronesecret",
            "registry": "ghcr.io",

        },
    }
    if test:
        step["settings"]["dry_run"] = True
    else:
        step["settings"]["auto_tag"] = True
    return step
