"""Regression tests for mdformat setup lock reclamation."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import unittest
from pathlib import Path
from unittest.mock import patch

REPO_ROOT = Path(__file__).resolve().parent.parent
SUPERVISOR_PATH = REPO_ROOT / "mise-tasks/setup/mdformat-supervisor"


def load_supervisor_module():
    loader = importlib.machinery.SourceFileLoader("mdformat_supervisor", str(SUPERVISOR_PATH))
    spec = importlib.util.spec_from_loader(loader.name, loader)
    if spec is None:
        raise RuntimeError(f"create import spec for {SUPERVISOR_PATH}")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


class LockReclamationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.supervisor_type = load_supervisor_module().Supervisor

    def test_ownerless_lock_stays_during_grace_period(self):
        reclaimable = self.supervisor_type.ownerless_lock_is_reclaimable(60)

        self.assertFalse(
            reclaimable,
            "an ownerless lock must remain during the 60-second owner-record creation grace period",
        )

    def test_ownerless_lock_becomes_reclaimable_after_grace_period(self):
        reclaimable = self.supervisor_type.ownerless_lock_is_reclaimable(61)

        self.assertTrue(
            reclaimable,
            "an ownerless lock older than 60 seconds must remain recoverable",
        )

    def test_transient_process_lookup_failure_does_not_report_dead_owner(self):
        supervisor = self.supervisor_type.__new__(self.supervisor_type)
        observations = iter((False, True))
        supervisor._process_matches = lambda _pid, _started: next(observations)

        with patch("time.sleep"):
            confirmed_dead = supervisor.owner_is_confirmed_dead(123, "start-id")

        self.assertFalse(
            confirmed_dead,
            "one transient process lookup failure must not permit a waiter to steal "
            "the live owner's lock",
        )

    def test_three_process_lookup_misses_confirm_dead_owner(self):
        supervisor = self.supervisor_type.__new__(self.supervisor_type)
        supervisor._process_matches = lambda _pid, _started: False

        with patch("time.sleep"):
            confirmed_dead = supervisor.owner_is_confirmed_dead(123, "start-id")

        self.assertTrue(
            confirmed_dead,
            "three consecutive process lookup misses must preserve prompt crash recovery",
        )


if __name__ == "__main__":
    unittest.main()
