import axios from 'axios'

const http = axios.create({
  baseURL: '/api',
  timeout: 30000
})

http.interceptors.response.use(
  res => res.data,
  err => {
    const msg = err.response?.data?.message || err.message
    return Promise.reject(new Error(msg))
  }
)

// ── 命名空间 ──────────────────────────────────────────
export const getNamespaces = () => http.get('/namespaces')

// ── Pod ──────────────────────────────────────────────
export const getPods = (namespace) =>
  http.get('/pods', { params: { namespace } })

export const deletePod = (namespace, name) =>
  http.delete(`/pods/${namespace}/${name}`)

export const restartPod = (namespace, name) =>
  http.post(`/pods/${namespace}/${name}/restart`)

export const getPodLogs = (namespace, name, tail = 200, container = '') =>
  http.get(`/pods/${namespace}/${name}/logs`, {
    params: { tail, ...(container ? { container } : {}) }
  })

export const getPodEvents = (namespace, name) =>
  http.get(`/pods/${namespace}/${name}/events`)

// ── Deployment ───────────────────────────────────────
export const getDeployments = (namespace) =>
  http.get('/deployments', { params: { namespace } })

export const scaleDeployment = (namespace, name, replicas) =>
  http.patch(`/deployments/${namespace}/${name}/scale`, { replicas })

export const restartDeployment = (namespace, name) =>
  http.post(`/deployments/${namespace}/${name}/restart`)

// 镜像更新：{ container: "nginx", image: "nginx:1.26" }
export const updateImage = (namespace, name, container, image) =>
  http.patch(`/deployments/${namespace}/${name}/image`, { container, image })

// ── Service ──────────────────────────────────────────
export const getServices = (namespace) =>
  http.get('/services', { params: { namespace } })

// ── Node ─────────────────────────────────────────────
export const getNodes = () => http.get('/nodes')

// ── ConfigMap ────────────────────────────────────────
export const getConfigMaps = (namespace) =>
  http.get('/configmaps', { params: { namespace } })

export const getConfigMapDetail = (namespace, name) =>
  http.get(`/configmaps/${namespace}/${name}`)

export const updateConfigMap = (namespace, name, data) =>
  http.put(`/configmaps/${namespace}/${name}`, { data })

// ── Secret ───────────────────────────────────────────
export const getSecrets = (namespace) =>
  http.get('/secrets', { params: { namespace } })

export const getSecretDetail = (namespace, name) =>
  http.get(`/secrets/${namespace}/${name}`)

// ── Ingress ──────────────────────────────────────────
export const getIngresses = (namespace) =>
  http.get('/ingresses', { params: { namespace } })

// ── StatefulSet ──────────────────────────────────────
export const getStatefulSets = (namespace) =>
  http.get('/statefulsets', { params: { namespace } })

export const scaleStatefulSet = (namespace, name, replicas) =>
  http.patch(`/statefulsets/${namespace}/${name}/scale`, { replicas })
