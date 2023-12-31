-- MIT License
--
-- (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
--
-- Permission is hereby granted, free of charge, to any person obtaining a
-- copy of this software and associated documentation files (the "Software"),
-- to deal in the Software without restriction, including without limitation
-- the rights to use, copy, modify, merge, publish, distribute, sublicense,
-- and/or sell copies of the Software, and to permit persons to whom the
-- Software is furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included
-- in all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
-- THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
-- OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
-- ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
-- OTHER DEALINGS IN THE SOFTWARE.

-- Install hmsds schema version 13

BEGIN;

ALTER TABLE hwinv_by_loc
ADD COLUMN "parent_node" VARCHAR(63) NOT NULL DEFAULT '';

CREATE OR REPLACE VIEW hwinv_by_loc_with_partition AS
SELECT
    hwinv_by_loc.id             AS  "id",
    hwinv_by_loc.type           AS  "type",
    hwinv_by_loc.ordinal        AS  "ordinal",
    hwinv_by_loc.status         AS  "status",
    hwinv_by_loc.location_info  AS  "location_info",
    hwinv_by_loc.fru_id         AS  "fru_id",
    hwinv_by_fru.type           AS  "fru_type",
    hwinv_by_fru.subtype        AS  "fru_subtype",
    hwinv_by_fru.fru_info       AS  "fru_info",
    part_info.name              AS  "partition"
FROM hwinv_by_loc
LEFT JOIN hwinv_by_fru ON hwinv_by_loc.fru_id = hwinv_by_fru.fru_id
LEFT JOIN (
    SELECT
        component_group_members.component_id AS "id",
        component_groups.name                AS "name"
    FROM component_group_members
    LEFT JOIN component_groups ON component_group_members.group_id = component_groups.id
    WHERE component_group_members.group_namespace = '%%partition%%'
) AS part_info ON hwinv_by_loc.parent_node = part_info.id;

-- This updates the new hwinv_by_loc column. Components under nodes (processors, memory, etc)
-- get their parent_node set to the node above them. All others get id = parent_node. This
-- way parent_node can be used to associate components that can't be added to partitions with
-- the partition that their parent component is in.
CREATE OR REPLACE FUNCTION hwinv_by_loc_update_parents()
RETURNS VOID AS $$
DECLARE
    node_id RECORD;
BEGIN
    FOR node_id IN SELECT id FROM hwinv_by_loc WHERE type = 'Node' LOOP
        UPDATE hwinv_by_loc SET parent_node = node_id.id WHERE id SIMILAR TO node_id.id||'([[:alpha:]][[:alnum:]]*)?';
    END LOOP;
    UPDATE hwinv_by_loc SET parent_node = id WHERE parent_node = '';
END;
$$ LANGUAGE plpgsql;

SELECT hwinv_by_loc_update_parents();

-- Bump the schema version
insert into system values(0, 13, '{}'::JSON)
    on conflict(id) do update set schema_version=13;

COMMIT;
