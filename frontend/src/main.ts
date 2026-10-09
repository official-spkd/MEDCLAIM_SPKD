import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import { siapkanDemo } from "./api";
import "./styles.css";

siapkanDemo().finally(() => {
  createApp(App).use(router).mount("#app");
});
