from dataclasses import dataclass, field
from collections.abc import Sequence
from typing import Any
from langgraph.graph.message import add_messages
from langchain_core.messages import BaseMessage


@dataclass
class AgentState:
    messages: Sequence[BaseMessage] = field(default_factory=list)
    user_id: str = ""
    subscription_tier: str = "free"
    session_id: str = ""
    tool_results: dict[str, Any] = field(default_factory=dict)
