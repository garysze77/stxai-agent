"""LangGraph agent for financial analysis with tool calling."""
import logging
from langgraph.graph import StateGraph, END
from langgraph.prebuilt import ToolNode
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage
from llm.router import get_llm_client, get_fallback_client
from agent.prompts import SYSTEM_PROMPT
from agent.state import AgentState
from tools.stubs import ALL_TOOLS

logger = logging.getLogger(__name__)


def build_agent():
    primary_llm = get_llm_client()
    fallback_llm = get_fallback_client()
    model_with_tools = primary_llm.bind_tools(ALL_TOOLS)

    async def agent_node(state: AgentState) -> dict:
        messages = [SystemMessage(content=SYSTEM_PROMPT), *state.messages]

        # Try primary (Puter) first
        try:
            response = await model_with_tools.ainvoke(messages)
            return {"messages": [response]}
        except Exception as e:
            logger.warning(f"Primary LLM failed: {e}")

        # Fallback to MiniMax
        if fallback_llm is not None:
            try:
                fb_with_tools = fallback_llm.bind_tools(ALL_TOOLS)
                response = await fb_with_tools.ainvoke(messages)
                return {"messages": [response]}
            except Exception as e:
                logger.error(f"Fallback LLM failed: {e}")
                raise RuntimeError(f"All LLM providers failed: {e}")
        raise

    def should_continue(state: AgentState) -> str:
        last = state.messages[-1]
        if hasattr(last, "tool_calls") and last.tool_calls:
            return "tools"
        return END

    graph = StateGraph(AgentState)
    graph.add_node("agent", agent_node)
    graph.add_node("tools", ToolNode(ALL_TOOLS))
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")

    return graph.compile()


agent = build_agent()
