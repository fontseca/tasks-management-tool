CREATE OR REPLACE FUNCTION fetch_groups (
  IN p_owner_id  UUID,
  IN p_page      BIGINT,
  IN p_rpp       BIGINT,
  IN p_needle    TEXT,
  IN p_sort_expr TEXT
)
RETURNS SETOF "group"
LANGUAGE 'plpgsql'
AS $$
BEGIN
  CALL assert_user_exists (p_owner_id);
  CALL validate_rpp_and_page (p_rpp, p_page);
  CALL gen_search_pattern (p_needle);
  CALL validate_sort_expr (p_sort_expr);
  RETURN QUERY
        SELECT *
          FROM "group" g
         WHERE g."is_archived" IS FALSE
               AND lower (concat (g."name", ' ', g."description")) ~ p_needle
      ORDER BY (CASE WHEN p_sort_expr = '' THEN g."created_at" END) DESC,
               (CASE WHEN p_sort_expr = '+name' THEN g."name" END) ASC,
               (CASE WHEN p_sort_expr = '-name' THEN g."name" END) DESC,
               (CASE WHEN p_sort_expr = '+description' THEN g."description" END) ASC,
               (CASE WHEN p_sort_expr = '-description' THEN g."description" END) DESC
         LIMIT p_rpp
        OFFSET (p_rpp * (p_page - 1));
END;
$$;

ALTER FUNCTION fetch_groups
      OWNER TO "noda";
