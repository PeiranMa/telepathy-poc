apiVersion: batch/v1
kind: Job
metadata:
  name: bench-$ITEM
  labels:
    app: bench
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: bench
        image: r4aks.azurecr.io/stan-go:v2
        command: ["/go/bin/stan-bench"]
        args: ["-s", "nats://stan:6222", "-np", "5","-ns","5", "-n", "5000000","test-$ITEM"]
        resources:
          requests:
            cpu: "3"
            memory: "4Gi"
