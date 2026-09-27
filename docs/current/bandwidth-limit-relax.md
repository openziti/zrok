# Bandwidth-limit relaxation

The controller checks limited accounts after their bandwidth usage drops below the configured threshold. It restores each share's dial policy and removes the account's bandwidth-limit journal entry only when every applicable share succeeds. Failed accounts retain their entries for a later cycle, while other accounts continue. Policy lookup by deterministic name makes a partial retry safe.

On v2, public shares can have no `frontend_selection`. The relax action records this as an incomplete share and keeps the account limited. Reconstructing that share's policy from its name mappings is a later repair step.

The cycle uses one database transaction. A store error ends the cycle because PostgreSQL aborts the transaction after a failed statement.
