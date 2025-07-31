/*
 * MIT License
 *
 * (C) Copyright [2021-2022] Hewlett Packard Enterprise Development LP
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
-- Adds pruning function for existing hardware history tables of redundant events.

BEGIN;

-- For each unique comp_id grab the earliest event to use as a 'base' event. Then
-- compare the other events (in time order) to the base event for matching fru_id.
-- If the fru_id matches, delete it. If the fru_id doesn't match, keep the event
-- and set it as the new base event. This removes all except the 1st occurrence of
-- a hardware history event indicating a change.
CREATE OR REPLACE FUNCTION hwinv_hist_prune()
RETURNS VOID AS $$
DECLARE
    comp_id RECORD;
    fru_event1 RECORD;
    fru_event2 RECORD;
BEGIN
    FOR comp_id IN SELECT distinct id FROM hwinv_hist LOOP
        SELECT * INTO fru_event1 FROM hwinv_hist WHERE id = comp_id.id ORDER BY timestamp ASC LIMIT 1;
        FOR fru_event2 IN SELECT * FROM hwinv_hist WHERE id = comp_id.id AND timestamp != fru_event1.timestamp ORDER BY timestamp ASC LOOP
            -- TODO: Shouldn't this IF statement also check that the
            --       event_type is the same?  Otherwise we'll be pruning
            --       out unique events that haven't repeated?
            IF fru_event2.fru_id = fru_event1.fru_id THEN
                DELETE FROM hwinv_hist WHERE id = fru_event2.id AND fru_id = fru_event2.fru_id AND timestamp = fru_event2.timestamp;
            ELSE
                fru_event1 = fru_event2;
            END IF;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Bump the schema version
insert into system values(0, 18, '{}'::JSON)
    on conflict(id) do update set schema_version=18;

COMMIT;
