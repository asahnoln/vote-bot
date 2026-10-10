# NOTE: Cannot be killed by SIGINT (CTRL+C)
# TODO: Solve SIGINT problem
firestore:
  gcloud emulators firestore start --host-port=[::1]:8711
