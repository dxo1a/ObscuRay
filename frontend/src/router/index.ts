import { createRouter, createWebHistory } from "vue-router"
import HomePage from "../pages/HomePage.vue"

const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/', component: HomePage }
    ]
})

// router.beforeEach(async (to, from, next) => {

// })

export default router