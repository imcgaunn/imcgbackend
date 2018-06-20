/*create the main index table.
*/
CREATE TABLE blogposts (ID INTEGER PRIMARY KEY ASC,
post_s3_loc text,
post_meta_s3_loc text,
created_time date);

/* general form for inserting an index entry */
INSERT INTO blogposts (post_s3_loc, post_meta_s3_loc, created_time)
VALUES('', '', 'Tue Jun 19 20:42:44 EDT 2018');

/* retrieve an index entry by id */
SELECT * from blogposts where ID=?;

/* delete an index entry by id */
DELETE from blogposts where ID=?;
