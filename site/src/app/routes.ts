import { createBrowserRouter } from "react-router";
import { Layout } from "./components/layout";
import { HomePage } from "./components/home-page";
import { HowItWorksPage } from "./components/how-it-works-page";
import { DashboardPage } from "./components/dashboard-page";
import { DocsPage } from "./components/docs-page";
import { RunsPage } from "./components/runs-page";
import { RunDetailPage } from "./components/run-detail-page";
import { PromptsPage } from "./components/prompts-page";
import { PromptDetailPage } from "./components/prompt-detail-page";
import { EvalDetailPage } from "./components/eval-detail-page";

export const router = createBrowserRouter([
  {
    path: "/",
    Component: Layout,
    children: [
      { index: true, Component: HomePage },
      { path: "how-it-works", Component: HowItWorksPage },
      { path: "dashboard", Component: DashboardPage },
      { path: "docs", Component: DocsPage },
      { path: "runs", Component: RunsPage },
      { path: "runs/:runId", Component: RunDetailPage },
      { path: "prompts", Component: PromptsPage },
      { path: "prompts/:promptId", Component: PromptDetailPage },
      { path: "runs/:runId/eval/:promptId/:configName", Component: EvalDetailPage },
    ],
  },
]);
