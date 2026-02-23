<template>
  <div class="keystore-config">
    <el-form-item label="密钥类型" required>
      <el-select v-model="config.type" placeholder="请选择密钥类型" @change="onTypeChange">
        <el-option label="引用已有密钥 (name)" value="named" />
        <el-option label="直接指定文件 (file)" value="file" />
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

    <!-- file 类型：填写文件路径 -->
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
        <el-form-item label="签名密钥路径" required>
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
}

interface KeystoreConfig {
  type: 'named' | 'file'
  name?: string
  params: Record<string, string>
}

// 必填标识文本
const requiredText = computed(() => props.required ? ' (必填)' : '')

const props = defineProps<Props>()
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

// 根据协议类型过滤 keystore 列表
const filteredKeystores = computed(() => {
  const type = props.isTlcp ? 'tlcp' : 'tls'
  return props.keystores.filter((ks: any) => ks.type === type)
})
</script>

<style scoped>
.keystore-config {
  margin-bottom: 20px;
}
</style>
