#!/usr/bin/env python3
"""Apply GoWA status@broadcast patch to vendored whatsmeow send.go."""
from __future__ import annotations

import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: patch_whatsmeow_status_send.py <path/to/send.go>", file=sys.stderr)
        return 2

    send_go = Path(sys.argv[1])
    c = send_go.read_text(encoding="utf-8")
    if "StatusBroadcastParticipants" in c:
        return 0

    needle_struct = "AdditionalNodes *[]waBinary.Node\n}"
    repl_struct = (
        "AdditionalNodes *[]waBinary.Node\n"
        "\t// StatusBroadcastParticipants overrides status privacy resolution when sending to status@broadcast.\n"
        "\tStatusBroadcastParticipants []types.JID\n"
        "}"
    )
    if needle_struct not in c:
        print("send.go: could not find SendRequestExtra tail (upstream layout changed?)", file=sys.stderr)
        return 1
    c = c.replace(needle_struct, repl_struct, 1)

    old = """\t\t} else {
\t\t\tgroupParticipants, err = cli.getBroadcastListParticipants(ctx, to)
\t\t\tif err != nil {
\t\t\t\terr = fmt.Errorf("failed to get broadcast list members: %w", err)
\t\t\t\treturn
\t\t\t}
\t\t}"""

    new = """\t\t} else if len(req.StatusBroadcastParticipants) > 0 && to == types.StatusBroadcastJID {
\t\t\tgroupParticipants = append([]types.JID(nil), req.StatusBroadcastParticipants...)
\t\t\townIDNonAD := cli.getOwnID().ToNonAD()
\t\t\tif ownIDNonAD.IsEmpty() {
\t\t\t\terr = ErrNotLoggedIn
\t\t\t\treturn
\t\t\t}
\t\t\tselfIndex := -1
\t\t\tfor i, participant := range groupParticipants {
\t\t\t\tif participant.User == ownIDNonAD.User {
\t\t\t\t\tselfIndex = i
\t\t\t\t\tbreak
\t\t\t\t}
\t\t\t}
\t\t\tif selfIndex < 0 {
\t\t\t\tgroupParticipants = append(groupParticipants, ownIDNonAD)
\t\t\t}
\t\t} else {
\t\t\tgroupParticipants, err = cli.getBroadcastListParticipants(ctx, to)
\t\t\tif err != nil {
\t\t\t\terr = fmt.Errorf("failed to get broadcast list members: %w", err)
\t\t\t\treturn
\t\t\t}
\t\t}"""

    if old not in c:
        print("send.go: could not find broadcast participants block (upstream layout changed?)", file=sys.stderr)
        return 1
    c = c.replace(old, new, 1)
    send_go.write_text(c, encoding="utf-8", newline="\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
