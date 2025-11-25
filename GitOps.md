build and deploy a scalable ecommerce-inspired microservice app, fully automated from code commit to production deployment.
using Kafka and Mongo db written in go lang. With Order, payment, and Fulfillment services use clean code and clean architecture and Domain Driven design use the outbox pattern for Integration Events that are triggered by Domain events Orders domain should be manager of the SAGA and should the order state should be updated after payment received events and ordershipped events Each microservice should subscribe to the events that need to process the order and to orders should subscribe to the events that update the state if any event is a failure the state should be updated and compensation operations should be implemented. Use Change Data Capture Debezium for the outbox pattern after the data for the microservice is saved in the events table that is data is saved. 

Create a plan first
Then break the work up into smaller chunks by layer and sub task for layer internals 
Then provide Guidelines for developer agents with prompts 

Create a Requirements.md, Plan.md, Task.md, and Guidelines.md 

Once all the required markdowns are created and the plan is reviewed we will build the solution 

Use the below to help with the Requirements.md 

Coding Standards will and Acceptance criteria will be needed for each layer and Task and subtask and be part of the guidelines a human developer will cut and paste the prompts in the guidelines files and be responsible for making sure the task are completed 

Use sequence task and then concurrent where possible to allow for asynchronous development after the baseline dependencies are in place allow for concurrent work with sub task so multiple agents and work at the same time on smaller task 

🔧 Tools & Technologies:
Azure AKS Kubernetes → for container orchestration
Docker → for containerization
GitHub Actions → for CI/CD automation
GitHub → for source control and deployment manifests
Trivy → and SonarQube for container security and code quality
OWASP → for dependency vulnerability analysis
Helm → for packaging and deploying chart
ArgoCD → to enable GitOps and continuous deployment
Prometheus & Grafana → for monitoring and visualization
Go Lang

💻 Build:
🔸Provision an AKS cluster with scalable node groups using , kubectl, and Azure CLI and Terraform

🔸If needed Install and configured Docker, SonarQube, Trivy, Helm, ArgoCD, Prometheus, and Grafana on the server environment

🔸Create github actions pipeline that
 - Pulls code from GitHub
 - Performs dependency and security scans (OWASP, SonarQube, Trivy)
 - Builds Docker images and pushes them to the registry after passing quality gates

🔸Create GitHub Actions CD workflow that
 - Auto-updates the Kubernetes deployment manifests (deployment.yaml) with new Docker image tags
 - Pushes these updated manifests back to GitHub

🔸Integrat ArgoCD for GitOps-based continuous deployment so that any update to manifests triggers seamless, zero-downtime pod replacement with new images

🔸Deploy Prometheus & Grafana via Helm for real-time monitoring and dashboard visualizations

🔸Configured github actions to send email notifications for build and deployment statuses

What I need
Real-world end-to-end DevOps workflows, including CI/CD and automated cloud-native deployments
Importance of shifting security left using container image scanning and static code analysis
How GitOps streamlines deployments and increases reliability
The crucial role of observability with Prometheus & Grafana for maintaining production stability

🔰 Outcome:
This project demonstrates how to build, automate, secure, and scale microservices on Kubernetes using open-source tools and cloud infrastructure. Perfect for streaming, e-commerce, or any scalable app scenarios.