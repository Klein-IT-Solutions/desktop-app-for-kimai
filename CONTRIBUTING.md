# Contributing

Thanks for helping to improve Kimai Tracker!

## Issues

- **Bugs:** describe what you did, what you expected and what happened. Add the version
  (*Settings → Info*), distribution, desktop (GNOME, KDE …) and session type (X11/Wayland).
- **Ideas:** explain the use case – what you want to achieve – not only the solution.

## Pull requests

1. Fork the repository and create a branch from `main`.
2. Set up the build as described in the [README](README.md#build-from-source-ubuntu).
3. Keep changes focused; one topic per pull request.
4. Run the checks before pushing:

   ```bash
   make lint
   make test
   ```

   Changes to the Kimai client should also pass `make test-integration` against the local
   test instance (`make kimai-up`).
5. Describe what changed and how you tested it.

Code style: `gofmt` for Go; for the frontend follow the existing Svelte/Tailwind style.
After changing Go services, regenerate the TypeScript bindings (`make build` or `make dev`
does this automatically) and commit them.

By contributing you agree that your contributions are licensed under the [MIT License](LICENSE).
