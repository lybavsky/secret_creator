Sample k8s autocreater of secrets in new namespaces operator.

Creates SecretsCreator CDR. 
After creating new namespace this simple operator iterates over SecretsCreator CDR and creates secrets

Example:
-
```
---
apiVersion: api.ash.lt/v1
kind: SecretCreator
metadata:
  name: secretcreator-sample-docker-2
spec:
  secretName: ourdockersecret2
  secret:
    type: kubernetes.io/dockerconfigjson
    data:
      .dockerconfigjson: eyJhdXRocyI6eyJzb21lcmVnaXN0cnkuZG9tYWluLmNvbSI6eyJ1c2VybmFtZSI6ImxvZ2luIiwicGFzc3dvcmQiOiJzb21lcGFzc3dvcmQiLCJhdXRoIjoiYkc5bmFXNDZjMjl0WlhCaGMzTjNiM0prIn19fQ==
 ```

After create new namespace it will create secret with name "ourdockersecret2" and content from "secret field"
"secret" field is core/v1/secret structure

To install apply to your cluster directory dist:
```
kubectl apply -f dist
```

Operator creates namespace secret-creator-system