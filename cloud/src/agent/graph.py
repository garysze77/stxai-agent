"""Financial AI agent with tool-calling loop.

Uses a manual tool loop with LLMRouter for Puter→MiniMax automatic fallback.
"""
import json
import logging
from langgraph.graph import StateGraph, END
from langchain_core.messages import (
    SystemMessage, HumanMessage, AIMessage, ToolMessage,
)
from llm.router import LLMRouter
from agent.prompts import SYSTEM_PROMPT
from agent.state import AgentState
from tools.stubs import TOOLS_BY_NAME

logger = logging.getLogger(__name__)
MAX_TOOL_ITERATIONS = 5


async def _run_tool_loop(messages: list, tools: list, router: LLMRouter):
    """Run the tool-calling loop until the model returns a text response."""
    for _ in range(MAX_TOOL_ITERATIONS):
        response, source = await router.ainvoke(messages, tools)
        messages.append(response)

        if not hasattr(response, "tool_calls") or not response.tool_calls:
            return messages

        for tc in response.tool_calls:
            tool_name = tc.get("name", "")
            tool_args = tc.get("args", {})
            tool_id = tc.get("id", "")

            if tool_name in TOOLS_BY_NAME:
                try:
                    result = TOOLS_BY_NAME[tool_name].invoke(tool_args)
                    content = json.dumps(result, ensure_ascii=False)
                except Exception as e:
                    content = json.dumps({"error": str(e)})
            else:
                content = json.dumps({"error": f"Unknown tool: {tool_name}"})

            messages.append(ToolMessage(content=content, tool_call_id=tool_id))

    return messages


def build_agent():
    router = LLMRouter()
    tools = list(TOOLS_BY_NAME.values())

    async def agent_node(state: AgentState) -> dict:
        messages = [SystemMessage(content=SYSTEM_PROMPT), *state.messages]

        try:
            messages = await _run_tool_loop(messages, tools, router)
        except Exception as e:
            logger.error(f"All LLMs failed: {e}")
            return {"messages": [
                AIMessage(content=f"I encountered an error: {e}. Please try again or rephrase your request.")
            ]}

        return {"messages": messages[1:]}  # strip system message

    graph = StateGraph(AgentState)
    graph.add_node("agent", agent_node)
    graph.set_entry_point("agent")
    graph.add_edge("agent", END)

    return graph.compile()


agent = build_agent()
