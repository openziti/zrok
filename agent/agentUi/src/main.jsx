import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import AgentUi from "./AgentUi.jsx";

createRoot(document.getElementById('root')).render(
  <StrictMode>
      <AgentUi />
  </StrictMode>,
)
