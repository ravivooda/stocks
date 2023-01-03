# Deployment
Today, we use deploy our app to DigitalOcean for budget constraints

The high level overview of architecture includes.

1. [Github action - Deploy To DO](deploy_to_do.yml) triggers upon master push (as well as scheduled for everyday evening)
   * Runs app with `setup` configuration
   * Tests app with `serve` configuration for a few minutes
   * Builds a docker image
   * Deploys the docker image to DO
2. DO app (seal-app) is configured for automatic deployment for any image push.
   * The app spec for the DO app is configured for rolling out the image 