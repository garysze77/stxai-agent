"""Parse the Signal Analyst's markdown output into structured fields."""
import re
from dataclasses import dataclass, field


@dataclass
class ParsedSignal:
    directional_bias: str = ""
    confidence_score: int = 0
    signal_strength: str = ""
    resistance: str = ""
    support: str = ""
    data_quality: str = ""
    bull_drivers: list[str] = field(default_factory=list)
    bear_risks: list[str] = field(default_factory=list)
    catalyst_watch: str = ""


def parse_signal_card(text: str) -> ParsedSignal:
    """Extract structured signal fields from the Signal Analyst's markdown block."""
    s = ParsedSignal()

    # Directional Bias
    m = re.search(r'\*\*Directional Bias\*\*[:\s]*([^\n]+)', text)
    if m:
        s.directional_bias = m.group(1).strip()

    # Confidence Score
    m = re.search(r'\*\*Confidence Score\*\*[:\s]*(\d+)', text)
    if m:
        s.confidence_score = int(m.group(1))

    # Signal Strength
    m = re.search(r'\*\*Signal Strength\*\*[:\s]*([^\n]+)', text)
    if m:
        s.signal_strength = m.group(1).strip()

    # Key Levels
    m = re.search(r'Resistance:\s*\$?([^\n]+)', text)
    if m:
        s.resistance = m.group(1).strip()
    m = re.search(r'Support:\s*\$?([^\n]+)', text)
    if m:
        s.support = m.group(1).strip()

    # Data Quality
    m = re.search(r'\*\*Data Quality\*\*[:\s]*([^\n]+)', text)
    if m:
        s.data_quality = m.group(1).strip()

    # Bull Case Drivers (list items after "Bull Case Drivers")
    s.bull_drivers = _extract_list_items(text, r'Bull Case Drivers')
    s.bear_risks = _extract_list_items(text, r'Bear Case Risks')

    # Catalyst Watch
    m = re.search(r'\*\*Catalyst Watch\*\*[:\s]*([^\n]+)', text)
    if m:
        s.catalyst_watch = m.group(1).strip()

    return s


def _extract_list_items(text: str, section: str) -> list[str]:
    """Extract bullet points from a section."""
    items = []
    pattern = rf'{section}.*?\n((?:\s*-[^\n]+\n?)+)'
    m = re.search(pattern, text, re.IGNORECASE)
    if m:
        for line in m.group(1).strip().split('\n'):
            line = line.strip()
            if line.startswith('-'):
                items.append(line[1:].strip())
    return items[:3]
