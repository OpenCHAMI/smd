/*
 * MIT License
 *
 * (C) Copyright [2025] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

-- Removes duplicate "Detected" events from the hardware history table

CREATE OR REPLACE FUNCTION hwinv_hist_remove_duplicate_detected_events()
RETURNS VOID AS $$
BEGIN
    -- Creating this index speeds up execution by several orders of magnitude

    CREATE INDEX IF NOT EXISTS hwinvhist_id_ts_idx ON hwinv_hist (id, "timestamp");

    -- Run the pruning logic

    WITH ordered AS (
        -- Build a temporary view of all events, ordered by time per device (id)
        SELECT ctid, id, "timestamp", event_type,
                -- For each event, get the previous event type for the same id
                LAG(event_type) OVER (PARTITION BY id ORDER BY "timestamp") AS prev_type
        FROM hwinv_hist
        WHERE id IN (
            -- Limit to CPUs and GPUs only
            SELECT loc.id
            FROM hwinv_by_loc loc
            WHERE loc.type IN ('Processor', 'NodeAccel')
        )
    ),
    dups AS (
        -- Identify rows where both this and previous event are "Detected" for the same id
        SELECT ctid
        FROM ordered
        WHERE event_type = 'Detected' AND prev_type = 'Detected'
    )

    -- Now delete the rows that have been identified as duplicates

    DELETE FROM hwinv_hist
    WHERE ctid IN (SELECT ctid FROM dups);
END;
$$ LANGUAGE plpgsql;

-- Execute the pruning function

SELECT hwinv_hist_remove_duplicate_detected_events();

-- A full vacuum must be run to reclaim space but cannot run from a migration.
-- The cray-smd-init service will run it manually after the migration completes.

-- Bump the schema version
insert into system values(0, 21, '{}'::JSON)
    on conflict(id) do update set schema_version=21;
