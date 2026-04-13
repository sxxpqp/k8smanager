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

// container 参数可选，不传则取第一个容器
export const getPodLogs = (namespace, name, tail = 200, container = '') =>
  http.get(`/pods/${namespace}/${name}/logs`, {
    params: { tail, ...(container ? { container } : {}) }
  })

// Pod 事件
export const getPodEvents = (namespace, name) =>
  http.get(`/pods/${namespace}/${name}/events`)

// ── Deployment ───────────────────────────────────────
export const getDeployments = (namespace) =>
  http.get('/deployments', { params: { namespace } })

export const scaleDeployment = (namespace, name, replicas) =>
  http.patch(`/deployments/${namespace}/${name}/scale`, { replicas })

export const restartDeployment = (namespace, name) =>
  http.post(`/deployments/${namespace}/${name}/restart`)

// ── Service ──────────────────────────────────────────
export const getServices = (namespace) =>
  http.get('/services', { params: { namespace } })
