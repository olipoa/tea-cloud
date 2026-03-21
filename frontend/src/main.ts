import { createPinia } from "pinia";
import "tdesign-vue-next/es/style/index.css";
import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./styles/main.scss";

const app = createApp(App);

app.use(createPinia());
app.use(router);

app.mount("#app");
