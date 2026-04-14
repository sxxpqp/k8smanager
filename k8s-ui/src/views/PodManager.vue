<template>
  <div class="pod-manager">

    <!-- 统计卡片 -->
    <div class="stat-cards">
      <div class="stat-card total">
        <div class="stat-num">{{ pods.length }}</div>
        <div class="stat-label">全部 Pod</div>
      </div>
      <div class="stat-card running">
        <div class="stat-num">{{ stats.running }}</div>
        <div class="stat-label">Running</div>
      </div>
      <div class="stat-card pending">
        <div class="stat-num">{{ stats.pending }}</div>
        <div class="stat-label">Pending</div>
      </div>
      <div class="stat-card failed">
        <div class="stat-num">{{ stats.failed }}</div>
        <div class="stat-label">Failed</div>
      </div>
      <div class="stat-card deploy">
        <div class="stat-num">{{ deployments.length }}</div>
        <div class="stat-label">Deployments</div>
      </div>
      <div class="stat-card svc">
        <div class="stat-num">{{ services.length }}</div>
        <div class="stat-label">Services</div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <el-select v-model="namespace" @change="onNsChange" style="width:180px">
        <template #prefix><el-icon><Grid /></el-icon></template>
        <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
      </el-select>

      <el-input
        v-model="searchText"
        placeholder="搜索名称..."
        clearable
        style="width:200px"
        :prefix-icon="Search"
      />

      <el-select
        v-if="activeTab === 'pods'"
        v-model="statusFilter"
        style="width:130px"
        clearable
        placeholder="全部状态"
      >
        <el-option label="Running"   value="Running" />
        <el-option label="Pending"   value="Pending" />
        <el-option label="Failed"    value="Failed" />
        <el-option label="Succeeded" value="Succeeded" />
      </el-select>

      <el-divider direction="vertical" />

      <el-switch v-model="autoRefresh" @change="toggleAutoRefresh" active-text="自动刷新" inline-prompt />
      <el-select v-model="refreshInterval" style="width:80px" size="small"
        :disabled="!autoRefresh" @change="toggleAutoRefresh(autoRefresh)">
        <el-option label="15s" :value="15" />
        <el-option label="30s" :value="30" />
        <el-option label="60s" :value="60" />
      </el-select>

      <el-button :icon="Refresh" @click="loadAll" :loading="loading" circle />
      <span class="last-refresh" v-if="lastRefreshTime">{{ lastRefreshTime }} 更新</span>
    </div>

    <!-- Tabs -->
    <el-tabs v-model="activeTab" class="main-tabs">

      <!-- ══ Pod ══ -->
      <el-tab-pane name="pods">
        <template #label>
          Pods&nbsp;<el-badge :value="filteredPods.length" :max="999" type="primary" />
        </template>

        <el-table
          :data="filteredPods"
          v-loading="loading"
          stripe highlight-current-row
          style="width:100%" row-key="name"
          @row-click="openDetail"
        >
          <el-table-column width="36" align="center">
            <template #default="{ row }">
              <el-icon :color="statusColor(row.status)" :size="13">
                <component :is="statusIcon(row.status)" />
              </el-icon>
            </template>
          </el-table-column>

          <el-table-column label="Pod 名称" prop="name" min-width="220">
            <template #default="{ row }">
              <span class="mono primary">{{ row.name }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.status)" size="small" effect="light">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="就绪" width="65" align="center" prop="ready" />

          <el-table-column label="重启" width="65" align="center">
            <template #default="{ row }">
              <el-tag size="small" effect="plain"
                :type="row.restarts > 5 ? 'danger' : row.restarts > 0 ? 'warning' : 'success'">
                {{ row.restarts }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="节点" prop="node" width="140">
            <template #default="{ row }">
              <span class="muted">{{ row.node || '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }">
              <span class="muted">{{ row.age }}</span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <el-button size="small" :icon="Document"
                @click.stop="handleLogs(row)">日志</el-button>

              <el-button size="small" :icon="Monitor"
                @click.stop="handleTerminal(row)">终端</el-button>

              <el-button size="small" :icon="Warning"
                @click.stop="handleEvents(row)">事件</el-button>

              <el-button size="small" type="warning" :icon="RefreshRight"
                :loading="row._restarting"
                @click.stop="handleRestartPod(row)">重启</el-button>

              <el-popconfirm :title="`确认删除 ${row.name}?`"
                confirm-button-type="danger" confirm-button-text="删除"
                @confirm="handleDeletePod(row)">
                <template #reference>
                  <el-button size="small" type="danger" :icon="Delete"
                    :loading="row._deleting" @click.stop>删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ Deployment ══ -->
      <el-tab-pane name="deployments">
        <template #label>
          Deployments&nbsp;<el-badge :value="filteredDeployments.length" :max="999" type="primary" />
        </template>

        <el-table :data="filteredDeployments" v-loading="loading" stripe style="width:100%">
          <el-table-column label="名称" prop="name" min-width="200">
            <template #default="{ row }">
              <span class="mono">{{ row.name }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" effect="light"
                :type="row.replicas===0 ? 'info' : row.readyReplicas===row.replicas ? 'success' : 'warning'">
                {{ row.replicas===0 ? '已停止' : row.readyReplicas===row.replicas ? '正常' : '更新中' }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="镜像" prop="image" min-width="220">
            <template #default="{ row }">
              <el-tooltip :content="row.image" placement="top">
                <span class="mono muted small">{{ row.image }}</span>
              </el-tooltip>
            </template>
          </el-table-column>

          <el-table-column label="Pod 数量" width="190" align="center">
            <template #default="{ row }">
              <div class="scale-control">
                <el-button circle size="small" :icon="Minus"
                  :disabled="row.replicas<=0 || row._scaling"
                  @click="handleScale(row, row.replicas-1)" />
                <div class="scale-info">
                  <span class="scale-num" :class="{ stopped: row.replicas===0 }">{{ row.replicas }}</span>
                  <span class="scale-ready">{{ row.readyReplicas }}/{{ row.replicas }} 就绪</span>
                </div>
                <el-button circle size="small" :icon="Plus"
                  :disabled="row._scaling"
                  @click="handleScale(row, row.replicas+1)" />
              </div>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="240" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.replicas===0" size="small" type="success"
                :icon="VideoPlay" :loading="row._scaling"
                @click="handleScale(row,1)">启动</el-button>
              <el-button v-else size="small" type="warning"
                :icon="VideoPause" :loading="row._scaling"
                @click="handleScale(row,0)">停止</el-button>

              <el-button size="small" :icon="RefreshRight"
                :loading="row._restarting"
                @click="handleRestartDeploy(row)">重启</el-button>

              <el-button size="small" type="primary" :icon="Upload"
                @click="handleUpdateImage(row)">更新镜像</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ Service ══ -->
      <el-tab-pane name="services">
        <template #label>
          Services&nbsp;<el-badge :value="filteredServices.length" :max="999" type="primary" />
        </template>

        <el-table :data="filteredServices" v-loading="loading" stripe style="width:100%">
          <el-table-column label="名称" prop="name" min-width="200">
            <template #default="{ row }">
              <span class="mono primary">{{ row.name }}</span>
            </template>
          </el-table-column>

          <el-table-column label="类型" width="130">
            <template #default="{ row }">
              <el-tag size="small" effect="plain" :type="svcTagType(row.type)">{{ row.type }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="ClusterIP" prop="clusterIP" width="140">
            <template #default="{ row }">
              <span class="mono muted">{{ row.clusterIP }}</span>
            </template>
          </el-table-column>

          <el-table-column label="端口" prop="ports" min-width="180">
            <template #default="{ row }">
              <el-tag
                v-for="p in row.ports" :key="p"
                size="small" type="info" effect="plain"
                style="margin:2px; font-family:monospace"
              >{{ p }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="Selector" min-width="200">
            <template #default="{ row }">
              <el-tag
                v-for="(v,k) in row.selector" :key="k"
                size="small" effect="plain"
                style="margin:2px; font-family:monospace"
              >{{ k }}={{ v }}</el-tag>
              <span v-if="!row.selector || !Object.keys(row.selector).length" class="muted">-</span>
            </template>
          </el-table-column>

          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }">
              <span class="muted">{{ row.age }}</span>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ Node ══ -->
      <el-tab-pane name="nodes" lazy>
        <template #label>节点</template>

        <el-table :data="nodes" v-loading="nodesLoading" stripe style="width:100%">
          <el-table-column label="名称" prop="name" min-width="160">
            <template #default="{ row }">
              <span class="mono primary">{{ row.name }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status==='Ready'?'success':'danger'" size="small" effect="light">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="角色" width="130">
            <template #default="{ row }">
              <el-tag v-for="r in row.roles" :key="r" size="small" effect="plain"
                style="margin:2px">{{ r }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="IP" prop="ip" width="140">
            <template #default="{ row }">
              <span class="mono muted">{{ row.ip }}</span>
            </template>
          </el-table-column>

          <el-table-column label="CPU" prop="cpu" width="70" align="center" />
          <el-table-column label="内存" prop="memory" width="100" align="center" />

          <el-table-column label="系统" prop="os" min-width="160">
            <template #default="{ row }">
              <span class="muted small">{{ row.os }}</span>
            </template>
          </el-table-column>

          <el-table-column label="运行时" prop="runtime" min-width="160">
            <template #default="{ row }">
              <span class="mono muted small">{{ row.runtime }}</span>
            </template>
          </el-table-column>

          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }">
              <span class="muted">{{ row.age }}</span>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ ConfigMap ══ -->
      <el-tab-pane name="configmaps" lazy>
        <template #label>ConfigMaps</template>

        <el-table :data="filteredConfigMaps" v-loading="loading" stripe style="width:100%"
          @row-click="openConfigMap">
          <el-table-column label="名称" prop="name" min-width="220">
            <template #default="{ row }">
              <span class="mono primary">{{ row.name }}</span>
            </template>
          </el-table-column>

          <el-table-column label="命名空间" prop="namespace" width="140" />

          <el-table-column label="键数量" width="90" align="center">
            <template #default="{ row }">
              <el-tag type="info" size="small">{{ row.dataCount }} 个键</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }">
              <span class="muted">{{ row.age }}</span>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" :icon="Edit"
                @click.stop="openConfigMap(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ Secret ══ -->
      <el-tab-pane name="secrets" lazy>
        <template #label>Secrets</template>
        <el-table :data="filteredSecrets" v-loading="loading" stripe style="width:100%">
          <el-table-column label="名称" prop="name" min-width="220">
            <template #default="{ row }">
              <span class="mono primary">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column label="类型" prop="type" min-width="200">
            <template #default="{ row }">
              <el-tag size="small" effect="plain" type="info">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="键数量" width="90" align="center">
            <template #default="{ row }">
              <el-tag size="small">{{ row.keyCount }} 个键</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }"><span class="muted">{{ row.age }}</span></template>
          </el-table-column>
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button size="small" :icon="View" @click="openSecret(row)">查看</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ Ingress ══ -->
      <el-tab-pane name="ingresses" lazy>
        <template #label>Ingress</template>
        <el-table :data="filteredIngresses" v-loading="loading" stripe style="width:100%">
          <el-table-column label="名称" prop="name" min-width="180">
            <template #default="{ row }">
              <span class="mono primary">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column label="IngressClass" prop="className" width="140">
            <template #default="{ row }">
              <el-tag size="small" effect="plain">{{ row.className || '-' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="规则" min-width="300">
            <template #default="{ row }">
              <div v-for="rule in row.rules" :key="rule.host" class="ingress-rule">
                <span class="ingress-host">{{ rule.host }}</span>
                <span v-for="path in rule.paths" :key="path"
                  class="ingress-path muted small">{{ path }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }"><span class="muted">{{ row.age }}</span></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- ══ StatefulSet ══ -->
      <el-tab-pane name="statefulsets" lazy>
        <template #label>StatefulSets</template>
        <el-table :data="filteredStatefulSets" v-loading="loading" stripe style="width:100%">
          <el-table-column label="名称" prop="name" min-width="200">
            <template #default="{ row }"><span class="mono">{{ row.name }}</span></template>
          </el-table-column>
          <el-table-column label="状态" width="90">
            <template #default="{ row }">
              <el-tag size="small" effect="light"
                :type="row.replicas===0?'info':row.readyReplicas===row.replicas?'success':'warning'">
                {{ row.replicas===0?'已停止':row.readyReplicas===row.replicas?'正常':'更新中' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="镜像" prop="image" min-width="200">
            <template #default="{ row }">
              <span class="mono muted small">{{ row.image }}</span>
            </template>
          </el-table-column>
          <el-table-column label="Pod 数量" width="190" align="center">
            <template #default="{ row }">
              <div class="scale-control">
                <el-button circle size="small" :icon="Minus"
                  :disabled="row.replicas<=0||row._scaling"
                  @click="handleStsScale(row, row.replicas-1)" />
                <div class="scale-info">
                  <span class="scale-num" :class="{stopped:row.replicas===0}">{{ row.replicas }}</span>
                  <span class="scale-ready">{{ row.readyReplicas }}/{{ row.replicas }} 就绪</span>
                </div>
                <el-button circle size="small" :icon="Plus"
                  :disabled="row._scaling"
                  @click="handleStsScale(row, row.replicas+1)" />
              </div>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" prop="age" width="165">
            <template #default="{ row }"><span class="muted">{{ row.age }}</span></template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

    </el-tabs>

    <!-- Secret 查看弹窗 -->
    <SecretViewer
      v-model="secretViewerVisible"
      :namespace="secretTarget.namespace"
      :secret-name="secretTarget.name"
    />

    <!-- 镜像更新弹窗 -->
    <ImageUpdateDialog
      v-model="imageDialogVisible"
      :deployment="imageTarget"
      @updated="loadDeployments"
    />

    <!-- ConfigMap 编辑弹窗 -->
    <ConfigMapEditor
      v-model="cmEditorVisible"
      :namespace="cmTarget.namespace"
      :cm-name="cmTarget.name"
      @updated="loadConfigMaps"
    />

    <!-- 终端弹窗 -->
    <TerminalDialog
      v-model="termDialogVisible"
      :namespace="termTarget.namespace"
      :pod-name="termTarget.name"
      :containers="termTarget.containers"
    />

    <!-- 日志弹窗 -->
    <LogDialog
      v-model="logDialogVisible"
      :namespace="logTarget.namespace"
      :pod-name="logTarget.name"
      :containers="logTarget.containers"
    />

    <!-- 事件弹窗 -->
    <el-dialog
      v-model="eventsDialogVisible"
      :title="`事件 — ${eventsTarget.name}`"
      width="760px"
      top="6vh"
    >
      <div v-loading="eventsLoading" style="min-height:200px">
        <div v-if="!eventsLoading && !events.length" class="empty-tip">
          <el-icon :size="40" color="#ccc"><InfoFilled /></el-icon>
          <p>该 Pod 暂无事件</p>
        </div>

        <el-timeline v-else>
          <el-timeline-item
            v-for="(e, i) in events" :key="i"
            :type="e.type === 'Warning' ? 'danger' : 'primary'"
            :timestamp="e.lastTime"
            placement="top"
          >
            <el-card shadow="never" class="event-card">
              <div class="event-header">
                <el-tag
                  :type="e.type === 'Warning' ? 'danger' : 'success'"
                  size="small" effect="light"
                >{{ e.type }}</el-tag>
                <span class="event-reason">{{ e.reason }}</span>
                <el-tag v-if="e.count > 1" type="info" size="small">×{{ e.count }}</el-tag>
                <span class="muted small" style="margin-left:auto">首次: {{ e.firstTime }}</span>
              </div>
              <div class="event-msg">{{ e.message }}</div>
            </el-card>
          </el-timeline-item>
        </el-timeline>
      </div>

      <template #footer>
        <el-button @click="eventsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Pod 详情抽屉 -->
    <el-drawer v-model="detailVisible" :title="detailPod?.name" size="460px" direction="rtl">
      <template v-if="detailPod">
        <el-tabs>
          <!-- 基本信息 -->
          <el-tab-pane label="基本信息">
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="命名空间">{{ detailPod.namespace }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="statusTagType(detailPod.status)" size="small">{{ detailPod.status }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="就绪">{{ detailPod.ready }}</el-descriptions-item>
              <el-descriptions-item label="重启次数">
                <el-tag :type="detailPod.restarts>5?'danger':detailPod.restarts>0?'warning':'success'" size="small">
                  {{ detailPod.restarts }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="节点">{{ detailPod.node || '-' }}</el-descriptions-item>
              <el-descriptions-item label="创建时间">{{ detailPod.age }}</el-descriptions-item>
              <el-descriptions-item label="容器列表">
                <el-tag v-for="c in detailPod.containers" :key="c"
                  size="small" style="margin:2px; font-family:monospace">{{ c }}</el-tag>
              </el-descriptions-item>
            </el-descriptions>

            <el-divider>快捷操作</el-divider>
            <div style="display:flex; gap:8px; flex-wrap:wrap">
              <el-button type="primary" :icon="Document" @click="handleLogs(detailPod)">查看日志</el-button>
              <el-button type="warning" :icon="RefreshRight" @click="handleRestartPod(detailPod)">重启</el-button>
              <el-popconfirm :title="`确认删除 ${detailPod.name}?`"
                confirm-button-type="danger"
                @confirm="handleDeletePod(detailPod); detailVisible=false">
                <template #reference>
                  <el-button type="danger" :icon="Delete">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </el-tab-pane>

          <!-- 快速跳转事件 -->
          <el-tab-pane label="事件">
            <div style="padding:20px 0; text-align:center">
              <el-button type="primary" :icon="Warning"
                @click="handleEvents(detailPod); detailVisible=false">
                查看 {{ detailPod?.name }} 的事件
              </el-button>
            </div>
          </el-tab-pane>
        </el-tabs>
      </template>
    </el-drawer>

  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Refresh, RefreshRight, Delete, Document, Edit,
  VideoPlay, VideoPause, Plus, Minus, Upload,
  Search, Grid, Warning, InfoFilled, Monitor, View,
  SuccessFilled, WarningFilled, CircleCloseFilled, QuestionFilled
} from '@element-plus/icons-vue'
import {
  getNamespaces, getPods, deletePod, restartPod,
  getDeployments, scaleDeployment, restartDeployment,
  getServices, getPodEvents, getNodes, getConfigMaps,
  getSecrets, getIngresses, getStatefulSets, scaleStatefulSet
} from '../api/index.js'
import LogDialog         from '../components/LogDialog.vue'
import TerminalDialog    from '../components/TerminalDialog.vue'
import ImageUpdateDialog from '../components/ImageUpdateDialog.vue'
import ConfigMapEditor   from '../components/ConfigMapEditor.vue'
import SecretViewer      from '../components/SecretViewer.vue'

// ── 状态 ──────────────────────────────────────────────
const namespace       = ref('default')
const namespaces      = ref([])
const pods            = ref([])
const deployments     = ref([])
const services        = ref([])
const loading         = ref(false)
const activeTab       = ref('pods')
const searchText      = ref('')
const statusFilter    = ref('')
const autoRefresh     = ref(false)
const refreshInterval = ref(30)
const lastRefreshTime = ref('')
let   refreshTimer    = null

// 日志
const logDialogVisible = ref(false)
const logTarget        = ref({ namespace: '', name: '', containers: [] })

// 终端
const termDialogVisible = ref(false)
const termTarget        = ref({ namespace: '', name: '', containers: [] })

// 节点
const nodes        = ref([])
const nodesLoading = ref(false)

// Secret
const secrets             = ref([])
const secretViewerVisible = ref(false)
const secretTarget        = ref({ namespace: '', name: '' })
const filteredSecrets     = computed(() =>
  searchText.value ? secrets.value.filter(s => s.name.includes(searchText.value)) : secrets.value
)

// Ingress
const ingresses        = ref([])
const filteredIngresses = computed(() =>
  searchText.value ? ingresses.value.filter(i => i.name.includes(searchText.value)) : ingresses.value
)

// StatefulSet
const statefulSets         = ref([])
const filteredStatefulSets = computed(() =>
  searchText.value ? statefulSets.value.filter(s => s.name.includes(searchText.value)) : statefulSets.value
)

// ConfigMap
const configMaps           = ref([])
const cmEditorVisible      = ref(false)
const cmTarget             = ref({ namespace: '', name: '' })
const filteredConfigMaps   = computed(() =>
  searchText.value
    ? configMaps.value.filter(c => c.name.includes(searchText.value))
    : configMaps.value
)

// 镜像更新
const imageDialogVisible = ref(false)
const imageTarget        = ref(null)

// 事件弹窗（独立）
const eventsDialogVisible = ref(false)
const eventsTarget        = ref({ namespace: '', name: '' })
const events              = ref([])
const eventsLoading       = ref(false)

// 详情抽屉
const detailVisible = ref(false)
const detailPod     = ref(null)

// ── 统计 ──────────────────────────────────────────────
const stats = computed(() => ({
  running: pods.value.filter(p => p.status === 'Running').length,
  pending: pods.value.filter(p => p.status === 'Pending').length,
  failed:  pods.value.filter(p => p.status === 'Failed').length,
}))

// ── 过滤 ──────────────────────────────────────────────
const filteredPods = computed(() => {
  let list = pods.value
  if (statusFilter.value) list = list.filter(p => p.status === statusFilter.value)
  if (searchText.value)   list = list.filter(p => p.name.includes(searchText.value))
  return list
})
const filteredDeployments = computed(() =>
  searchText.value
    ? deployments.value.filter(d => d.name.includes(searchText.value))
    : deployments.value
)
const filteredServices = computed(() =>
  searchText.value
    ? services.value.filter(s => s.name.includes(searchText.value))
    : services.value
)

// ── 状态样式 ──────────────────────────────────────────
const statusTagType = s => ({ Running:'success', Pending:'warning', Failed:'danger', Succeeded:'info' }[s] ?? 'info')
const statusColor   = s => ({ Running:'#67c23a', Pending:'#e6a23c', Failed:'#f56c6c' }[s] ?? '#909399')
const statusIcon    = s => ({ Running: SuccessFilled, Pending: WarningFilled, Failed: CircleCloseFilled }[s] ?? QuestionFilled)
const svcTagType    = t => ({ ClusterIP:'primary', NodePort:'success', LoadBalancer:'warning', ExternalName:'info' }[t] ?? 'info')

// ── 数据加载 ──────────────────────────────────────────
async function loadNamespaces() {
  try {
    const res = await getNamespaces()
    namespaces.value = res.data ?? res
  } catch (e) {
    ElMessage.error('获取命名空间失败: ' + e.message)
  }
}

async function loadPods() {
  try {
    const res = await getPods(namespace.value)
    pods.value = (res.data ?? res).map(p => ({ ...p, _restarting: false, _deleting: false }))
  } catch (e) {
    ElMessage.error('获取 Pod 失败: ' + e.message)
  }
}

async function loadDeployments() {
  try {
    const res = await getDeployments(namespace.value)
    const fresh = res.data ?? res
    if (deployments.value.length === 0) {
      deployments.value = fresh.map(d => ({ ...d, _scaling: false, _restarting: false }))
    } else {
      fresh.forEach(d => {
        const exist = deployments.value.find(r => r.name === d.name && r.namespace === d.namespace)
        if (exist) {
          if (!exist._scaling) exist.replicas = d.replicas
          exist.readyReplicas = d.readyReplicas
          exist.image = d.image
        } else {
          deployments.value.push({ ...d, _scaling: false, _restarting: false })
        }
      })
      deployments.value = deployments.value.filter(r =>
        fresh.some(d => d.name === r.name && d.namespace === r.namespace)
      )
    }
  } catch (e) {
    ElMessage.error('获取 Deployment 失败: ' + e.message)
  }
}

async function loadServices() {
  try {
    const res = await getServices(namespace.value)
    services.value = res.data ?? res
  } catch (e) {
    ElMessage.error('获取 Service 失败: ' + e.message)
  }
}

async function loadNodes() {
  nodesLoading.value = true
  try {
    const res = await getNodes()
    nodes.value = res.data ?? []
  } catch (e) {
    ElMessage.error('获取节点失败: ' + e.message)
  } finally {
    nodesLoading.value = false
  }
}

async function loadConfigMaps() {
  try {
    const res = await getConfigMaps(namespace.value)
    // 兼容 Go 返回大写字段（未加 json tag 时）
    configMaps.value = (res.data ?? []).map(cm => ({
      name:      cm.name      ?? cm.Name,
      namespace: cm.namespace ?? cm.Namespace ?? namespace.value,
      dataCount: cm.dataCount ?? cm.DataCount ?? 0,
      age:       cm.age       ?? cm.Age ?? ''
    }))
  } catch (e) {
    ElMessage.error('获取 ConfigMap 失败: ' + e.message)
  }
}

async function loadAll() {
  loading.value = true
  const tasks = [loadPods(), loadDeployments(), loadServices(), loadConfigMaps()]
  // 已经加载过的懒加载 Tab，刷新时一并更新
  if (nodes.value.length       || activeTab.value === 'nodes')        tasks.push(loadNodes())
  if (secrets.value.length     || activeTab.value === 'secrets')      tasks.push(loadSecrets())
  if (ingresses.value.length   || activeTab.value === 'ingresses')    tasks.push(loadIngresses())
  if (statefulSets.value.length|| activeTab.value === 'statefulsets') tasks.push(loadStatefulSets())
  await Promise.all(tasks)
  loading.value = false
  lastRefreshTime.value = new Date().toLocaleTimeString()
}

function onNsChange() {
  pods.value        = []
  deployments.value = []
  services.value    = []
  configMaps.value  = []
  secrets.value     = []
  ingresses.value   = []
  statefulSets.value= []
  // nodes 是集群级别，不受 namespace 影响，无需清空
  loadAll()
}

// Tab 懒加载
watch(activeTab, tab => {
  if (tab === 'nodes'       && !nodes.value.length)       loadNodes()
  if (tab === 'secrets'     && !secrets.value.length)     loadSecrets()
  if (tab === 'ingresses'   && !ingresses.value.length)   loadIngresses()
  if (tab === 'statefulsets'&& !statefulSets.value.length) loadStatefulSets()
})

function handleUpdateImage(row) {
  // 把 deployment 的容器信息传给弹窗
  // 后端需要在 deployment 列表里返回 containers: [{name, image}]
  imageTarget.value = {
    name:       row.name,
    namespace:  row.namespace,
    containers: row.containers ?? [{ name: 'app', image: row.image }]
  }
  imageDialogVisible.value = true
}

function openConfigMap(row) {
  cmTarget.value        = { namespace: row.namespace, name: row.name }
  cmEditorVisible.value = true
}

async function loadSecrets() {
  try {
    const res = await getSecrets(namespace.value)
    secrets.value = res.data ?? []
  } catch (e) { ElMessage.error('获取 Secret 失败: ' + e.message) }
}

function openSecret(row) {
  secretTarget.value        = { namespace: row.namespace, name: row.name }
  secretViewerVisible.value = true
}

async function loadIngresses() {
  try {
    const res = await getIngresses(namespace.value)
    ingresses.value = res.data ?? []
  } catch (e) { ElMessage.error('获取 Ingress 失败: ' + e.message) }
}

async function loadStatefulSets() {
  try {
    const res = await getStatefulSets(namespace.value)
    const fresh = res.data ?? []
    if (statefulSets.value.length === 0) {
      statefulSets.value = fresh.map(s => ({ ...s, _scaling: false }))
    } else {
      fresh.forEach(s => {
        const exist = statefulSets.value.find(r => r.name === s.name && r.namespace === s.namespace)
        if (exist) {
          if (!exist._scaling) exist.replicas = s.replicas
          exist.readyReplicas = s.readyReplicas
          exist.image = s.image
        } else {
          statefulSets.value.push({ ...s, _scaling: false })
        }
      })
      // 删除已不存在的
      statefulSets.value = statefulSets.value.filter(r =>
        fresh.some(s => s.name === r.name && s.namespace === r.namespace)
      )
    }
  } catch (e) { ElMessage.error('获取 StatefulSet 失败: ' + e.message) }
}

async function handleStsScale(row, replicas) {
  if (replicas < 0) return
  const prev = row.replicas
  try {
    row._scaling = true
    row.replicas = replicas
    await scaleStatefulSet(row.namespace, row.name, replicas)
    ElMessage.success(`已扩缩容至 ${replicas} 副本`)
    setTimeout(loadStatefulSets, 2000)
  } catch (e) {
    row.replicas = prev
    ElMessage.error(e.message)
  } finally {
    row._scaling = false
  }
}

// ── 自动刷新 ──────────────────────────────────────────
function toggleAutoRefresh(val) {
  clearInterval(refreshTimer)
  if (val) refreshTimer = setInterval(loadAll, refreshInterval.value * 1000)
}

// ── Pod 操作 ──────────────────────────────────────────
async function handleRestartPod(row) {
  try {
    row._restarting = true
    await restartPod(row.namespace, row.name)
    ElMessage.success(`${row.name} 重启中`)
    setTimeout(loadPods, 2000)
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    row._restarting = false
  }
}

async function handleDeletePod(row) {
  try {
    row._deleting = true
    await deletePod(row.namespace, row.name)
    ElMessage.success('Pod 已删除')
    await loadPods()
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    row._deleting = false
  }
}

function handleLogs(row) {
  logTarget.value = {
    namespace:  row.namespace,
    name:       row.name,
    containers: row.containers ?? []
  }
  logDialogVisible.value = true
}

function handleTerminal(row) {
  termTarget.value = {
    namespace:  row.namespace,
    name:       row.name,
    containers: row.containers ?? []
  }
  termDialogVisible.value = true
}

function openDetail(row) {
  detailPod.value     = row
  detailVisible.value = true
}

async function handleEvents(row) {
  eventsTarget.value        = { namespace: row.namespace, name: row.name }
  eventsDialogVisible.value = true
  events.value              = []
  eventsLoading.value       = true
  try {
    const res = await getPodEvents(row.namespace, row.name)
    events.value = res.data ?? []
  } catch (e) {
    ElMessage.error('获取事件失败: ' + e.message)
  } finally {
    eventsLoading.value = false
  }
}

// ── Deployment 操作 ───────────────────────────────────
async function handleScale(row, replicas) {
  if (replicas < 0) return
  const prev = row.replicas
  try {
    row._scaling = true
    row.replicas = replicas
    await scaleDeployment(row.namespace, row.name, replicas)
    ElMessage.success(`已扩缩容至 ${replicas} 副本`)
    setTimeout(loadDeployments, 2000)
  } catch (e) {
    row.replicas = prev
    ElMessage.error(e.message)
  } finally {
    row._scaling = false
  }
}

async function handleRestartDeploy(row) {
  try {
    row._restarting = true
    await restartDeployment(row.namespace, row.name)
    ElMessage.success(`${row.name} 滚动重启中`)
    setTimeout(loadAll, 3000)
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    row._restarting = false
  }
}

// ── 生命周期 ──────────────────────────────────────────
onMounted(async () => {
  await loadNamespaces()
  await loadAll()
})
onUnmounted(() => clearInterval(refreshTimer))
</script>

<style scoped>
.pod-manager {
  padding: 14px;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
}

/* 统计卡片 */
.stat-cards { display: flex; gap: 10px; }
.stat-card {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  padding: 12px 14px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  border-left: 4px solid #ddd;
}
.stat-card.total  { border-color: #409eff; }
.stat-card.running{ border-color: #67c23a; }
.stat-card.pending{ border-color: #e6a23c; }
.stat-card.failed { border-color: #f56c6c; }
.stat-card.deploy { border-color: #909399; }
.stat-card.svc    { border-color: #9b59b6; }
.stat-num   { font-size: 24px; font-weight: 700; color: #303133; line-height: 1; }
.stat-label { font-size: 12px; color: #909399; margin-top: 4px; }

/* 工具栏 */
.toolbar {
  display: flex; align-items: center; gap: 10px; flex-wrap: wrap;
  background: #fff; padding: 10px 14px;
  border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}
.last-refresh { font-size: 12px; color: #bbb; margin-left: auto; }

/* Tab 面板 */
.main-tabs {
  flex: 1; overflow: auto;
  background: #fff; padding: 0 16px 16px;
  border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

/* 文字辅助类 */
.mono    { font-family: 'Consolas','Monaco',monospace; font-size: 13px; }
.primary { color: #409eff; }
.muted   { color: #888; }
.small   { font-size: 12px; }

/* 扩缩容 */
.scale-control { display: flex; align-items: center; justify-content: center; gap: 10px; }
.scale-info    { display: flex; flex-direction: column; align-items: center; min-width: 52px; }
.scale-num     { font-size: 22px; font-weight: 700; color: #409eff; line-height: 1; }
.scale-num.stopped { color: #909399; }
.scale-ready   { font-size: 11px; color: #909399; margin-top: 2px; }

/* Ingress 规则 */
.ingress-rule  { margin-bottom: 4px; }
.ingress-host  { font-family: monospace; font-weight: 600; color: #409eff; margin-right: 8px; }
.ingress-path  { display: block; padding-left: 12px; }

/* Event 卡片 */
.empty-tip { text-align: center; color: #ccc; padding: 40px 0; }
.event-card { margin-bottom: 4px; }
.event-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.event-reason { font-weight: 600; font-size: 13px; }
.event-msg    { font-size: 12px; color: #555; line-height: 1.5; }
</style>
