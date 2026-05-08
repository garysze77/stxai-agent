"""STX AI Agent — multi-agent debate architecture.

Simple mode: single agent with tool loop (fast, cheap).
Deep mode: Bullish Analyst → Bearish Analyst → Lead Analyst (comprehensive).
"""
import json
import logging
from langgraph.graph import StateGraph, END
from langchain_core.messages import (
    SystemMessage, HumanMessage, AIMessage, ToolMessage,
)
from llm.router import LLMRouter
from agent.prompts import (
    SYSTEM_PROMPT, BULLISH_ANALYST_PROMPT,
    BEARISH_ANALYST_PROMPT, LEAD_ANALYST_PROMPT,
)
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
            return response.content if hasattr(response, "content") else str(response), source

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

    # hit max iterations — return last AI message content
    for msg in reversed(messages):
        if hasattr(msg, "content") and msg.content and not hasattr(msg, "tool_call_id"):
            return msg.content, "max_iterations"
    return "Analysis did not complete within tool iterations.", "max_iterations"


def _extract_last_ai(messages: list) -> str:
    for msg in reversed(messages):
        if isinstance(msg, AIMessage) and msg.content:
            return msg.content
    return ""


# ── Simple Agent (single LLM + tools, current behavior) ──


def build_simple_agent():
    router = LLMRouter()
    tools = list(TOOLS_BY_NAME.values())

    async def agent_node(state: AgentState) -> dict:
        messages = [SystemMessage(content=SYSTEM_PROMPT), *state.messages]
        try:
            messages = await _run_tool_loop_raw(messages, tools, router)
        except Exception as e:
            logger.error(f"Simple agent failed: {e}")
            return {"messages": [
                AIMessage(content=f"Error: {e}. Please try again.")
            ]}
        return {"messages": messages[1:]}

    graph = StateGraph(AgentState)
    graph.add_node("agent", agent_node)
    graph.set_entry_point("agent")
    graph.add_edge("agent", END)
    return graph.compile()


async def _run_tool_loop_raw(messages: list, tools: list, router: LLMRouter) -> list:
    """Run tool loop and return full messages list (for simple agent)."""
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


# ── Deep Agent (Multi-Agent Debate) ──


def build_deep_agent():
    router = LLMRouter()
    tools = list(TOOLS_BY_NAME.values())

    async def bullish_node(state: AgentState) -> dict:
        """Bullish Analyst: gather data + build bull case."""
        user_msg = state.messages[-1].content if state.messages else ""
        messages = [
            SystemMessage(content=BULLISH_ANALYST_PROMPT),
            HumanMessage(content=f"Build the strongest possible bull case for: {user_msg}"),
        ]
        try:
            thesis, source = await _run_tool_loop(messages, tools, router)
            logger.info(f"Bullish thesis from {source}, length={len(thesis)}")
        except Exception as e:
            logger.error(f"Bullish analyst failed: {e}")
            thesis = f"Bullish analysis unavailable: {e}"

        return {"bullish_thesis": thesis}

    async def bearish_node(state: AgentState) -> dict:
        """Bearish Analyst: read bull thesis, gather data, build bear case + rebuttal."""
        user_msg = state.messages[-1].content if state.messages else ""
        bull_thesis = state.bullish_thesis or "(No bullish thesis provided)"

        messages = [
            SystemMessage(content=BEARISH_ANALYST_PROMPT),
            HumanMessage(content=f"""User request: {user_msg}

BULLISH THESIS (rebut this):
{bull_thesis}

Now build the strongest possible bear case. Directly counter the bullish arguments above."""),
        ]
        try:
            thesis, source = await _run_tool_loop(messages, tools, router)
            logger.info(f"Bearish thesis from {source}, length={len(thesis)}")
        except Exception as e:
            logger.error(f"Bearish analyst failed: {e}")
            thesis = f"Bearish analysis unavailable: {e}"

        return {"bearish_thesis": thesis}

    async def lead_node(state: AgentState) -> dict:
        """Lead Analyst: synthesize bull + bear theses into final report."""
        user_msg = state.messages[-1].content if state.messages else ""
        bull_thesis = state.bullish_thesis or "(No bullish thesis)"
        bear_thesis = state.bearish_thesis or "(No bearish thesis)"

        messages = [
            SystemMessage(content=LEAD_ANALYST_PROMPT),
            HumanMessage(content=f"""User request: {user_msg}

BULLISH THESIS:
{bull_thesis}

BEARISH THESIS:
{bear_thesis}

Synthesize these into the definitive STX AI analysis report. Weigh the evidence from both sides.""")
        ]
        try:
            response, source = await router.ainvoke(messages)  # no tools for synthesis
            reply = response.content if hasattr(response, "content") else str(response)
            logger.info(f"Lead synthesis from {source}, length={len(reply)}")
        except Exception as e:
            logger.error(f"Lead analyst failed: {e}")
            # fallback: combine both theses manually
            reply = f"""## Analysis

### 🟢 Bull Case
{bull_thesis[:800]}

### 🔴 Bear Case
{bear_thesis[:800]}

### ⚠️ Note
Synthesis unavailable due to an error. Please review both cases above."""

        return {"messages": [AIMessage(content=reply)]}

    graph = StateGraph(AgentState)
    graph.add_node("bullish_analyst", bullish_node)
    graph.add_node("bearish_analyst", bearish_node)
    graph.add_node("lead_analyst", lead_node)

    graph.set_entry_point("bullish_analyst")
    graph.add_edge("bullish_analyst", "bearish_analyst")
    graph.add_edge("bearish_analyst", "lead_analyst")
    graph.add_edge("lead_analyst", END)

    return graph.compile()


# Pre-built graphs
simple_agent = build_simple_agent()
deep_agent = build_deep_agent()
