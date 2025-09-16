<template>
    <div>
        <header class="flex justify-between items-center bg-gray-700" style="--wails-draggable: drag">
            <div class="w-fit flex gap-2 items-center ml-2">
                <img class="w-[16px] h-[16px] mt-2 mb-2" :src="isAnyProfileActive ? appIconActive : appIcon" />
                <label class="text-[14px]">ObscuRay</label>
            </div>
            <div class="flex items-center mr-1">
                <button @click="WindowMinimise" class="pr-1 text-gray-400 hover:text-gray-300">
                    <Minus :size="18" />
                </button>
                <button @click="Quit" class="p-2 text-gray-400 hover:text-red-600">
                    <X :size="18" />
                </button>
            </div>
        </header>
        <div class="p-4 flex flex-col h-[calc(104vh-50px)]">
            <!-- table -->
            <div
                class="w-full border border-gray-600 mb-4 rounded overflow-hidden flex flex-col flex-1"
                role="table"
            >
                <!-- columheaders -->
                <div class="bg-gray-700 flex" role="row">
                <div class="p-2 w-[10px] font-semibold select-none" role="columnheader">ID</div>
                <div class="p-2 flex-1 font-semibold select-none" role="columnheader">{{ $t('profileName') }}</div>
                <div class="p-2 flex-1 font-semibold select-none" role="columnheader">{{ $t('status') }}</div>
                <div class="p-2 flex-1 font-semibold select-none" role="columnheader">{{ $t('actions') }}</div>
                </div>

                <!-- scrollable -->
                <div class="flex-1 overflow-y-auto min-h-0">
                    <transition-group name="fade" tag="div">
                        <div
                            v-for="profile in profiles"
                            :key="profile.id"
                            class="flex hover:bg-gray-900 border-gray-600 items-center transition relative"
                            role="row"
                        >
                            <div class="pl-2.5 pr-2 w-[10px]" role="cell">{{ profile.id }}</div>
                            <div class="p-2 flex-1" role="cell">{{ profile.name }}</div>
                            <div class="p-2 flex-1 select-none" role="cell">
                            {{ profile.isActive ? $t('active') : $t('inactive') }}
                            </div>
                            <div class="p-2 flex-1" role="cell">
                                <button
                                    @click="toggleProfile(profile.id)"
                                    class="px-6 py-1.5 rounded items-center gap-1"
                                    :class="profile.isActive ? 'bg-red-500 hover:bg-red-600' : 'bg-blue-500 hover:bg-blue-600'"
                                >
                                    <SquarePause v-if="profile.isActive" :size="20" />
                                    <SquarePlay v-else :size="20" />
                                </button>
                                
                            </div>
                            <button
                                @click="deleteProfile(profile.id)"
                                class="px-1.5 py-1.5 rounded bg-gray-500 hover:bg-gray-600 disabled:bg-gray-700 disabled:text-gray-800 right-2.5 absolute"
                                :disabled="profile.isActive"
                            >
                                <Trash :size="20"/>
                            </button>
                        </div>
                    </transition-group>
                </div>
            </div>

            <!-- footer -->
            <div
                @contextmenu.prevent="addProfile"
                class="p-4 border rounded text-gray-600 hover:text-white border-gray-700 hover:border-gray-500 text-center cursor-pointer transition"
            >
                {{ $t('addProfileHint') }}
            </div>
            <div class="mt-4 text-right">
                <p class="text-xs select-none flex justify-end space-x-2">
                    <span>{{ downloadSpeed }} ↓</span>
                    <span>{{ uploadSpeed }} ↑</span>
                </p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { GetProfiles, AddProfileFromClipboard, StartProfile, StopProfile, DeleteProfile, GetTrafficStats } from "../../wailsjs/go/main/App.js"
import { WindowIsMinimised, Quit, WindowMinimise } from "../../wailsjs/runtime/runtime.js"
import { X, Minus, SquarePause, SquarePlay, Trash } from "lucide-vue-next"
import appIcon from "../assets/images/appicon3.png"
import appIconActive from "../assets/images/appicon3_1.png"

interface Profile {
    id: string
    name: string
    vless: string
    isActive: boolean
}

const profiles = ref<Profile[]>([])
const downloadSpeed = ref('0 B/s')
const uploadSpeed = ref('0 B/s')
let statsInterval: number | null = null

const isAnyProfileActive = computed(() => profiles.value.some(p => p.isActive))

const loadProfiles = async () => {
    try {
        profiles.value = await GetProfiles()
    } catch (err) {
        alert(`Error load profiles: ${err}`)
    }
}

const addProfile = async () => {
    try {
        await AddProfileFromClipboard()
        await loadProfiles()
    } catch (err) {
        alert(`Error adding profile: ${err}`)
    }
}

const toggleProfile = async (id: string) => {
    try {
        const currentProfile = profiles.value.find(p => p.id === id)
        if (currentProfile?.isActive) {
            await StopProfile()
        } else {
            await StartProfile(id)
        }
        await loadProfiles()
    } catch (err) {
        alert(`Error toggling profile: ${err}`)
    }
}

const deleteProfile = async (id: string) => {
    try {
        await DeleteProfile(id)
        await loadProfiles()
    } catch (err) {
        alert(`Error profile deletion: ${err}`)
    }
}

const updateStats = async () => {
    try {
        if (!profiles.value.some(p => p.isActive)) {
            downloadSpeed.value = '0 B/s'
            uploadSpeed.value = '0 B/s'
            return
        }
        const stats = await GetTrafficStats()
        downloadSpeed.value = formatSpeed(stats.download ?? 0)
        uploadSpeed.value = formatSpeed(stats.upload ?? 0)
    } catch (err) {
        console.error('Error getting stats:', err)
        downloadSpeed.value = '0 B/s'
        uploadSpeed.value = '0 B/s'
    }
}

const formatSpeed = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B/s`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB/s`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB/s`
}

onMounted(() => {
    loadProfiles()
    statsInterval = setInterval(async () => {
        const isMinimised = await WindowIsMinimised()
        if (!isMinimised) {
            await updateStats()
        }
    }, 2000)
})

onUnmounted(() => {
    if (statsInterval) {
        clearInterval(statsInterval)
    }
})
</script>
