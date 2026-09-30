<template>
  <div class="keystore-config">
    <el-form-item label="密钥类型" required>
      <el-select v-model="config.type" placeholder="请选择密钥类型" @change="onTypeChange">
        <el-option label="引用已有密钥 (named)" value="named" />
        <el-option v-if="!ibc" label="直接指定文件 (file)" value="file" />
        <el-option v-if="ibc" label="直接指定 IBC 材料文件 (ibc-file)" value="ibc-file" />
      </el-select>
    </el-form-item>

    <!-- name 类型：选择已有 keystore -->
    <template v-if="config && config.type === 'named'">
      <el-form-item :label="'密钥名称' + requiredText" :required="required">
        <el-select v-model="config.name" placeholder="请选择密钥" clearable>
          <el-option v-for="ks in filteredKeystores" :key="ks.name" :label="ks.name" :value="ks.name" />
        </el-select>
      </el-form-item>
    </template>

    <!-- file 类型：填写证书密钥文件路径 -->
    <template v-if="config && config.type === 'file'">
      <el-form-item :label="'签名证书路径' + requiredText" :required="required">
        <el-input v-model="config.params['sign-cert']" placeholder="./keystores/sign.crt" />
      </el-form-item>
      <el-form-item :label="'签名密钥路径' + requiredText" :required="required">
        <el-input v-model="config.params['sign-key']" placeholder="./keystores/sign.key" />
      </el-form-item>
      <template v-if="isTlcp">
        <el-form-item label="加密证书路径" required>
          <el-input v-model="config.params['enc-cert']" placeholder="./keystores/enc.crt" />
        </el-form-item>
        <el-form-item label="加密密钥路径" required>
          <el-input v-model="config.params['enc-key']" placeholder="./keystores/enc.key" />
        </el-form-item>
      </template>
    </template>

    <!-- ibc-file 类型：填写 IBC 身份材料文件路径（标识、KGC 公共参数、三把用户私钥） -->
    <template v-if="config && config.type === 'ibc-file'">
      <el-form-item :label="'标识文件路径' + requiredText" :required="required">
        <el-input v-model="config.params['identity']" placeholder="./keystores/ibc-identity.txt" />
      </el-form-item>
      <el-form-item :label="'KGC 公共参数路径' + requiredText" :required="required">
        <el-input v-model="config.params['params']" placeholder="./keystores/ibc-params.pem" />
      </el-form-item>
      <el-form-item label="签名私钥路径 (hid=0x01)">
        <el-input v-model="config.params['sign-key']" placeholder="./keystores/ibc-sign.key" />
      </el-form-item>
      <el-form-item label="加密私钥路径 (hid=0x03)">
        <el-input v-model="config.params['enc-key']" placeholder="./keystores/ibc-enc.key" />
      </el-form-item>
      <el-form-item label="密钥交换私钥路径 (hid=0x02)">
        <el-input v-model="config.params['kex-key']" placeholder="./keystores/ibc-kex.key" />
      </el-form-item>
      <el-alert
        title="签名私钥用于签名/验签，加密私钥用于解密预主密钥，密钥交换私钥用于 SM9 密钥交换（IBSDH 套件必需），三者不可混用。"
        type="info"
        :closable="false"
        style="margin-bottom: 16px;"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  keystores: any[]
  isTlcp: boolean
  modelValue?: KeystoreConfig
  required?: boolean
  // 是否为 IBC 身份 keystore 配置块：下拉提供 ibc-file，并只过滤 ibc 类型的 keystore
  ibc?: boolean
}

interface KeystoreConfig {
  type: 'named' | 'file' | 'ibc-file'
  name?: string
  params: Record<string, string>
}

// 必填标识文本
const props = withDefaults(defineProps<Props>(), {
  ibc: false,
})
const requiredText = computed(() => props.required ? ' (必填)' : '')

const emit = defineEmits<{
  'update:modelValue': [value: KeystoreConfig]
}>()

const config = computed({
  get: () => props.modelValue || { type: 'named', name: undefined, params: {} },
  set: (value) => emit('update:modelValue', value)
})

function onTypeChange() {
  emit('update:modelValue', {
    type: config.value.type,
    name: undefined,
    params: {}
  })
}

// 根据协议/上下文过滤 keystore 列表：IBC 块仅列出 ibc 类型，证书块按协议列出 tlcp/tls
const filteredKeystores = computed(() => {
  if (props.ibc) {
    return props.keystores.filter((ks: any) => ks.type === 'ibc')
  }
  const type = props.isTlcp ? 'tlcp' : 'tls'
  return props.keystores.filter((ks: any) => ks.type === type)
})
</script>

<style scoped>
.keystore-config {
  margin-bottom: 20px;
}
</style>