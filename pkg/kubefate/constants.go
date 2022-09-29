// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubefate

// GetDefaultYAML returns the default kubefate deployment yaml
func GetDefaultYAML() string {
	return defaultKubeFATEYAMLVersion180
}

const defaultKubeFATEYAMLVersion180 = `{{- if .IsClusterAdmin -}}
apiVersion: v1
kind: Namespace
metadata:
  name: kube-fate
  labels:
    name: kube-fate
---
{{ end -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubefate-admin
  namespace: {{.Namespace}}
---
apiVersion: v1
kind: Secret
metadata:
  name: kubefate-secret
  namespace: {{.Namespace}}
type: Opaque
stringData:
  kubefateUsername: {{.ServiceUserName}}
  kubefatePassword: {{.ServicePassword}}
  mariadbUsername: kubefate
  mariadbPassword: kubefate
{{- if .IsClusterAdmin}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kubefate
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kubefate-role
subjects:
  - kind: ServiceAccount
    name: kubefate-admin
    namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubefate-role
  namespace: kube-fate
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  - configmaps
  - services
  - secrets
  - persistentvolumeclaims
  - serviceaccounts
  verbs:
  - get
  - list
  - create
  - delete
  - update
  - patch
- apiGroups:
  - ""
  resources:
  - pods
  - pods/log
  - nodes
  verbs:
  - get
  - list
- apiGroups:
  - apps
  resources:
  - deployments
  - statefulsets
  verbs:
  - get
  - list
  - create
  - delete
  - update
  - patch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - create
  - delete
  - update
  - patch
- apiGroups:
  - networking.istio.io
  resources:
  - gateway
  - virtualservice
  verbs:
  - get
  - create
  - delete
  - update
  - patch
- apiGroups:
  - policy
  resources:
  - podsecuritypolicies
  verbs:
  - get
  - use
  - create
  - delete
  - update
  - patch
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - roles
  - rolebindings
  verbs:
  - get
  - create
  - delete
  - update
  - patch
{{- else }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: sc-edit-binding
  namespace: {{.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: edit
subjects:
  - kind: ServiceAccount
    name: kubefate-admin
    namespace: {{.Namespace}}
{{- end }}
{{- if .UseImagePullSecrets}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.ImagePullSecretsName}}
  namespace: {{.Namespace}}
data:
  .dockerconfigjson: {{.RegistrySecretData}}
type: kubernetes.io/dockerconfigjson
{{- end }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubefate
  namespace: {{.Namespace}}
  labels:
    fate: kubefate
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      fate: kubefate
  template:
    metadata:
      labels:
        fate: kubefate
    spec:
      serviceAccountName: kubefate-admin
      containers:
        {{- if .UseRegistry}}
        - image: {{.Registry}}/kubefate:v1.4.5
        {{- else }}
        - image: federatedai/kubefate:v1.4.5
        {{- end }}
          imagePullPolicy: IfNotPresent
          name: kubefate
          env:
            - name: FATECLOUD_DB_TYPE
              value: "mysql"
            - name: FATECLOUD_DB_HOST
              value: "mariadb"
            - name: FATECLOUD_DB_PORT
              value: "3306"
            - name: FATECLOUD_DB_NAME
              value: "kube_fate"
            - name: FATECLOUD_DB_USERNAME
              valueFrom:
                secretKeyRef:
                  name: kubefate-secret
                  key: mariadbUsername
            - name: FATECLOUD_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kubefate-secret
                  key: mariadbPassword
            - name: FATECLOUD_REPO_NAME
              value: "kubefate"
            - name: FATECLOUD_REPO_URL
              value: "https://federatedai.github.io/KubeFATE"
            - name: FATECLOUD_USER_USERNAME
              valueFrom:
                secretKeyRef:
                  name: kubefate-secret
                  key: kubefateUsername
            - name: FATECLOUD_USER_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kubefate-secret
                  key: kubefatePassword
            - name: FATECLOUD_SERVER_ADDRESS
              value: "0.0.0.0"
            - name: FATECLOUD_SERVER_PORT
              value: "8080"
            - name: FATECLOUD_LOG_LEVEL
              value: "debug"
            - name: FATECLOUD_LOG_NOCOLOR
              value: "true"
          ports:
            - containerPort: 8080
          resources:
            limits:
              memory: "4Gi"
              cpu: "2"
            requests:
              memory: 512Mi
              cpu: "0.5"
          livenessProbe:
            tcpSocket:
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 3
          readinessProbe:
            tcpSocket:
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 1
            successThreshold: 1
            failureThreshold: 3
          startupProbe:
            tcpSocket:
              port: 8080
            failureThreshold: 30
            periodSeconds: 10
      {{- if .UseImagePullSecrets}}
      imagePullSecrets:
        - name: {{.ImagePullSecretsName}}
      {{- end }}
      restartPolicy: Always
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mariadb
  namespace: {{.Namespace}}
  labels:
    fate: mariadb
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      fate: mariadb
  template:
    metadata:
      labels:
        fate: mariadb
    spec:
      serviceAccountName: kubefate-admin
      containers:
        {{- if .UseRegistry}}
        - image: {{.Registry}}/mariadb:10
        {{- else }}
        - image: mariadb:10
        {{- end }}
          imagePullPolicy: IfNotPresent
          name: mariadb
          env:
            - name: MYSQL_DATABASE
              value: "kube_fate"
            - name: MYSQL_ALLOW_EMPTY_PASSWORD
              value: "1"
            - name: MYSQL_USER
              valueFrom:
                secretKeyRef:
                  name: kubefate-secret
                  key: mariadbUsername
            - name: MYSQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: kubefate-secret
                  key: mariadbPassword
          ports:
            - containerPort: 3306
          resources:
            limits:
              memory: "4Gi"
              cpu: "2"
            requests:
              memory: 512Mi
              cpu: "0.5"
          volumeMounts:
            - name: mariadb-data
              mountPath: /var/lib/mysql
          livenessProbe:
            exec:
              command:
              - mysqladmin
              - ping
            initialDelaySeconds: 30
            periodSeconds: 30
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 3
          readinessProbe:
            exec:
              command:
              - mysqladmin
              - ping
            initialDelaySeconds: 30
            periodSeconds: 30
            timeoutSeconds: 1
            successThreshold: 1
            failureThreshold: 3
      restartPolicy: Always
      {{- if .UseImagePullSecrets}}
      imagePullSecrets:
        - name: {{.ImagePullSecretsName}}
      {{- end }}
      volumes:
        - name: mariadb-data
          # If you want persistence, please modify it to the PV you want to use.
          emptyDir: {}
          # hostPath:
          #   path: /data/kubefate/mysql/db
          #   type: DirectoryOrCreate
          # persistentVolumeClaim:
          #   claimName: "mariadb-data"
# ---
# apiVersion: v1
# kind: PersistentVolumeClaim
# metadata:
#   name: mariadb-data
#   namespace: {{.Namespace}}
# spec:
#   accessModes:
#   - ReadWriteOnce
#   storageClassName: <MyStorageClassName>
#   resources:
#     requests:
#       storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: mariadb
  namespace: {{.Namespace}}
  labels:
    fate: mariadb
spec:
  ports:
    - name: "3306"
      port: 3306
      targetPort: 3306
      protocol: TCP
  type: ClusterIP
  selector:
    fate: mariadb
---
apiVersion: v1
kind: Service
metadata:
  name: kubefate
  namespace: {{.Namespace}}
  labels:
    fate: kubefate
spec:
  ports:
    - name: "8080"
      port: 8080
      targetPort: 8080
      protocol: TCP
  type: ClusterIP
  selector:
    fate: kubefate
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubefate
  namespace: {{.Namespace}}
spec:
  ingressClassName: nginx
  rules:
    - host: {{.Hostname}}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service: 
                name: kubefate
                port: 
                  number: 8080`

// GetDefaultIngressControllerYAML returns the default ingress controller deployment yaml
func GetDefaultIngressControllerYAML() string {
	return defaultIngressNginxVersion111
}

const defaultIngressNginxVersion111 = `
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx

---
# Source: ingress-nginx/templates/controller-serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx
  namespace: ingress-nginx
automountServiceAccountToken: true
---
# Source: ingress-nginx/templates/controller-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx-controller
  namespace: ingress-nginx
data:
  allow-snippet-annotations: 'true'
---
# Source: ingress-nginx/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
  name: ingress-nginx
rules:
  - apiGroups:
      - ''
    resources:
      - configmaps
      - endpoints
      - nodes
      - pods
      - secrets
      - namespaces
    verbs:
      - list
      - watch
  - apiGroups:
      - ''
    resources:
      - nodes
    verbs:
      - get
  - apiGroups:
      - ''
    resources:
      - services
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ''
    resources:
      - events
    verbs:
      - create
      - patch
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingresses/status
    verbs:
      - update
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingressclasses
    verbs:
      - get
      - list
      - watch
---
# Source: ingress-nginx/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
  name: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ingress-nginx
subjects:
  - kind: ServiceAccount
    name: ingress-nginx
    namespace: ingress-nginx
---
# Source: ingress-nginx/templates/controller-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx
  namespace: ingress-nginx
rules:
  - apiGroups:
      - ''
    resources:
      - namespaces
    verbs:
      - get
  - apiGroups:
      - ''
    resources:
      - configmaps
      - pods
      - secrets
      - endpoints
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ''
    resources:
      - services
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingresses/status
    verbs:
      - update
  - apiGroups:
      - networking.k8s.io
    resources:
      - ingressclasses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ''
    resources:
      - configmaps
    resourceNames:
      - ingress-controller-leader
    verbs:
      - get
      - update
  - apiGroups:
      - ''
    resources:
      - configmaps
    verbs:
      - create
  - apiGroups:
      - ''
    resources:
      - events
    verbs:
      - create
      - patch
  - apiGroups:
      - policy
    resources:
      - podsecuritypolicies
    verbs:
      - use
---
# Source: ingress-nginx/templates/controller-rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx
  namespace: ingress-nginx
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ingress-nginx
subjects:
  - kind: ServiceAccount
    name: ingress-nginx
    namespace: ingress-nginx
---
# Source: ingress-nginx/templates/controller-service-webhook.yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx-controller-admission
  namespace: ingress-nginx
spec:
  type: ClusterIP
  ports:
    - name: https-webhook
      port: 443
      targetPort: webhook
      appProtocol: https
  selector:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/component: controller
---
# Source: ingress-nginx/templates/controller-service.yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  type: {{.ServiceType}}
  externalTrafficPolicy: Cluster
  ipFamilyPolicy: SingleStack
  ipFamilies:
    - IPv4
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: http
      appProtocol: http
    - name: https
      port: 443
      protocol: TCP
      targetPort: https
      appProtocol: https
  selector:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/component: controller
---
# Source: ingress-nginx/templates/controller-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: ingress-nginx
      app.kubernetes.io/instance: ingress-nginx
      app.kubernetes.io/component: controller
  revisionHistoryLimit: 10
  minReadySeconds: 0
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/component: controller
    spec:
      dnsPolicy: ClusterFirst
      containers:
        - name: controller
          image: k8s.gcr.io/ingress-nginx/controller:v1.1.1@sha256:0bc88eb15f9e7f84e8e56c14fa5735aaa488b840983f87bd79b1054190e660de
          imagePullPolicy: IfNotPresent
          lifecycle:
            preStop:
              exec:
                command:
                  - /wait-shutdown
          args:
            - /nginx-ingress-controller
            - --watch-ingress-without-class=true
            - --publish-service=$(POD_NAMESPACE)/ingress-nginx-controller
            - --election-id=ingress-controller-leader
            - --controller-class=k8s.io/ingress-nginx
            - --configmap=$(POD_NAMESPACE)/ingress-nginx-controller
            - --validating-webhook=:8443
            - --validating-webhook-certificate=/usr/local/certificates/cert
            - --validating-webhook-key=/usr/local/certificates/key
          securityContext:
            capabilities:
              drop:
                - ALL
              add:
                - NET_BIND_SERVICE
            runAsUser: 101
            allowPrivilegeEscalation: true
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: LD_PRELOAD
              value: /usr/local/lib/libmimalloc.so
          livenessProbe:
            failureThreshold: 5
            httpGet:
              path: /healthz
              port: 10254
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 10254
              scheme: HTTP
            initialDelaySeconds: 10
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          ports:
            - name: http
              containerPort: 80
              protocol: TCP
            - name: https
              containerPort: 443
              protocol: TCP
            - name: webhook
              containerPort: 8443
              protocol: TCP
          volumeMounts:
            - name: webhook-cert
              mountPath: /usr/local/certificates/
              readOnly: true
          resources:
            requests:
              cpu: 100m
              memory: 90Mi
      nodeSelector:
        kubernetes.io/os: linux
      serviceAccountName: ingress-nginx
      terminationGracePeriodSeconds: 300
      volumes:
        - name: webhook-cert
          secret:
            secretName: ingress-nginx-admission
---
# Source: ingress-nginx/templates/controller-ingressclass.yaml
# We don't support namespaced ingressClass yet
# So a ClusterRole and a ClusterRoleBinding is required
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: controller
  name: nginx
  namespace: ingress-nginx
spec:
  controller: k8s.io/ingress-nginx
---
# Source: ingress-nginx/templates/admission-webhooks/validating-webhook.yaml
# before changing this value, check the required kubernetes version
# https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
  name: ingress-nginx-admission
webhooks:
  - name: validate.nginx.ingress.kubernetes.io
    matchPolicy: Equivalent
    rules:
      - apiGroups:
          - networking.k8s.io
        apiVersions:
          - v1
        operations:
          - CREATE
          - UPDATE
        resources:
          - ingresses
    failurePolicy: Fail
    sideEffects: None
    admissionReviewVersions:
      - v1
    clientConfig:
      service:
        namespace: ingress-nginx
        name: ingress-nginx-controller-admission
        path: /networking/v1/ingresses
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ingress-nginx-admission
  namespace: ingress-nginx
  annotations:
    helm.sh/hook: pre-install,pre-upgrade,post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ingress-nginx-admission
  annotations:
    helm.sh/hook: pre-install,pre-upgrade,post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
rules:
  - apiGroups:
      - admissionregistration.k8s.io
    resources:
      - validatingwebhookconfigurations
    verbs:
      - get
      - update
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ingress-nginx-admission
  annotations:
    helm.sh/hook: pre-install,pre-upgrade,post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ingress-nginx-admission
subjects:
  - kind: ServiceAccount
    name: ingress-nginx-admission
    namespace: ingress-nginx
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: ingress-nginx-admission
  namespace: ingress-nginx
  annotations:
    helm.sh/hook: pre-install,pre-upgrade,post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
rules:
  - apiGroups:
      - ''
    resources:
      - secrets
    verbs:
      - get
      - create
  - apiGroups:
      - policy
    resources:
      - podsecuritypolicies
    verbs:
      - use
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ingress-nginx-admission
  namespace: ingress-nginx
  annotations:
    helm.sh/hook: pre-install,pre-upgrade,post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ingress-nginx-admission
subjects:
  - kind: ServiceAccount
    name: ingress-nginx-admission
    namespace: ingress-nginx
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/job-createSecret.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: ingress-nginx-admission-create
  namespace: ingress-nginx
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
spec:
  template:
    metadata:
      name: ingress-nginx-admission-create
      labels:
        helm.sh/chart: ingress-nginx-4.0.15
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/version: 1.1.1
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/component: admission-webhook
    spec:
      containers:
        - name: create
          image: k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1@sha256:64d8c73dca984af206adf9d6d7e46aa550362b1d7a01f3a0a91b20cc67868660
          imagePullPolicy: IfNotPresent
          args:
            - create
            - --host=ingress-nginx-controller-admission,ingress-nginx-controller-admission.$(POD_NAMESPACE).svc
            - --namespace=$(POD_NAMESPACE)
            - --secret-name=ingress-nginx-admission
          env:
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          securityContext:
            allowPrivilegeEscalation: false
      restartPolicy: OnFailure
      serviceAccountName: ingress-nginx-admission
      nodeSelector:
        kubernetes.io/os: linux
      securityContext:
        runAsNonRoot: true
        runAsUser: 2000
---
# Source: ingress-nginx/templates/admission-webhooks/job-patch/job-patchWebhook.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: ingress-nginx-admission-patch
  namespace: ingress-nginx
  annotations:
    helm.sh/hook: post-install,post-upgrade
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
  labels:
    helm.sh/chart: ingress-nginx-4.0.15
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/instance: ingress-nginx
    app.kubernetes.io/version: 1.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/component: admission-webhook
spec:
  template:
    metadata:
      name: ingress-nginx-admission-patch
      labels:
        helm.sh/chart: ingress-nginx-4.0.15
        app.kubernetes.io/name: ingress-nginx
        app.kubernetes.io/instance: ingress-nginx
        app.kubernetes.io/version: 1.1.1
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/component: admission-webhook
    spec:
      containers:
        - name: patch
          image: k8s.gcr.io/ingress-nginx/kube-webhook-certgen:v1.1.1@sha256:64d8c73dca984af206adf9d6d7e46aa550362b1d7a01f3a0a91b20cc67868660
          imagePullPolicy: IfNotPresent
          args:
            - patch
            - --webhook-name=ingress-nginx-admission
            - --namespace=$(POD_NAMESPACE)
            - --patch-mutating=false
            - --secret-name=ingress-nginx-admission
            - --patch-failure-policy=Fail
          env:
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
          securityContext:
            allowPrivilegeEscalation: false
      restartPolicy: OnFailure
      serviceAccountName: ingress-nginx-admission
      nodeSelector:
        kubernetes.io/os: linux
      securityContext:
        runAsNonRoot: true
        runAsUser: 2000`
