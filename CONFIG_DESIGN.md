# Configuration Design

This is just a little scratchpad / design document for figuring out how to deal
with configuration.


## Sources of Configuration

1. Local config file `flamenco-manager.yaml`
2. Environment variables (for easily putting into docker)
3. CLI parameters


## Flow of Configuration

1. Load at startup from `flamenco-manager.yaml`
    - Nice to have: monitoring & live reloading of that configuration file.
2. Load at startup from environment variables
    - Will never change.
3. Load at startup from CLI parameters
    - Will also never change
4. Receive new config via API (for Lineup integration)
    - Will require live adjustments of configuration.


## Design Questions
