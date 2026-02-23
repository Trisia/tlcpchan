<template>
  <div class="keystore-config">
    <el-form-item label="密钥类型" required>
      <el-select v-model="config.type" placeholder="请选择密钥类型" @change="onTypeChange">
        <el-option label="引用已有密钥 (name)" value="named" />
        <el-option label="直接指定文件 (file)" value="file" />
      </el-select>
    </el-form-item>

    <!-- name 类型：选择已有 keystore -->
    <template v-if="config.type === 'named'">
      <el-form-item label="密钥名称" required>
        <el-select v-model="config.name" placeholder="请选择密钥" clearable>
          <el-option
            v-for="ks in keystores"
            :key="ks.name"
            :label="ks.name"
            :value="ks.name"
          />
        </el-select>
      </el-form-item>
    </template>

    <!-- file 类型：填写文件路径 -->
    <template v-if="config.type === 'file'">
      <el-form-item label="签名证书路径" required>
        <el-input v-model="config.params['sign-cert']" placeholder="./keystores/sign.crt" />
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
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'

interface Props {
  keystores: any[]
  isTlcp: boolean
}

interface KeystoreConfig {
  type: 'named' | 'file'
  name?: string
  params: Record<string, string>
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: KeystoreConfig]
}>()

const config = reactive<KeystoreConfig>({
  type: 'named',
  name: undefined,
  params: {}
})

watch(
  () => config,
  (newVal) => {
    emit('update:modelValue', { ...newVal })
  },
  { deep: true }
)

function onTypeChange() {
  config.name = undefined
  config.params = {}
}
</script>

<style scoped>
.keystore-config {
  margin-bottom: 20px;
}
</style>
