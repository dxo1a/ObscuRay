<template>
    <div class="p-4 bg-gray-800 text-white min-h-screen">
        <h1 class="text-2xl mb-4">ObscuRay</h1>
        <table class="w-full border border-gray-600 mb-4">
            <thead>
                <tr class="bg-gray-700">
                    <th class="p-2">Имя профиля</th>
                    <th class="p-2">Статус</th>
                    <th class="p-2">Действия</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="profile in profiles" :key="profile.id" class="hover:bg-gray-600">
                    <td class="p-2">{{ profile.name }}</td>
                    <td class="p-2">{{ profile.isActive ? 'Активен' : 'Неактивен' }}</td>
                    <td class="p-2">
                        <button
                            @click="toggleProfile(profile.id)"
                            class="px-2 py-1 rounded text-white"
                            :class="profile.isActive ? 'bg-red-500 hover:bg-red-600' : 'bg-blue-500 hover:bg-blue-600'"
                        >
                            {{ profile.isActive ? 'Остановить' : 'Запустить' }}
                        </button>
                    </td>
                </tr>
            </tbody>
        </table>
        <div
            @contextmenu.prevent="addProfile"
            class="p-4 border border-dashed border-gray-400 text-center cursor-pointer"
        >
            ПКМ здесь для добавления VLESS из буфера обмена
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { GetProfiles, AddProfileFromClipboard, StartProfile, StopProfile } from '../../wailsjs/go/main/App.js'

interface Profile {
    id: string
    name: string
    vless: string
    isActive: boolean
}

const profiles = ref<Profile[]>([])

const loadProfiles = async () => {
    try {
        profiles.value = await GetProfiles()
    } catch (err) {
        alert(`Ошибка загрузки профилей: ${err}`)
    }
}

const addProfile = async () => {
    try {
        await AddProfileFromClipboard()
        await loadProfiles()
    } catch (err) {
        alert(`Ошибка добавления профиля: ${err}`)
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
        alert(`Ошибка управления профилем: ${err}`)
    }
}

onMounted(loadProfiles)
</script>
