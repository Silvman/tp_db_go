package mapper

import (
	"github.com/jackc/pgx"
	"io/ioutil"
)

//prepared
const qSelectForumBySlug = `select slug, title, posts, threads, owner from forums where slug = $1`
const qSelectUserByNick = `select nickname, fullname, about, email from users where nickname = $1`
const qSelectThreadById = `select id, title, message, votes, slug, created, forum, author from threads where id = $1`
const qSelectPostById = `select id, parent, message, isEdit, forum, created, thread, author from posts where id = $1`

const qSelectUsersSinceDesc = `
select nickname, fullname, about, email from forums_users where forum = $1 and nickname < $2
order by nickname desc
limit $3`

const qSelectUsersDesc = `
select nickname, fullname, about, email from forums_users where forum = $1
order by nickname desc
limit $2`

const qSelectUsersSince = `
select nickname, fullname, about, email from forums_users where forum = $1 and nickname > $2
order by nickname
limit $3`

const qSelectUsers = `
select nickname, fullname, about, email from forums_users where forum = $1
order by nickname
limit $2`

const qSelectThreadsCreatedDesc = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 and created <= $2::timestamptz order by created desc limit $3`
const qSelectThreadsCreated = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 and created >= $2::timestamptz order by created limit $3`
const qSelectThreadsDesc = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 order by created desc limit $2`
const qSelectThreads = `select id, title, message, votes, slug, created, forum, author from threads where forum = $1 order by created limit $2`

const qSelectIdFromThreadsId = `select id from threads where id = $1::bigint`
const qSelectIdFromThreadsSlug = `select id from threads where slug = $1`

const qSelectPostsPTDesc = `
with rootPosts as (
    select id from posts
    where thread = $1
      and parent = 0
    order by id desc
    limit $2
    ) select p.id, p.parent, p.message, p.isEdit, p.forum, p.created, p.thread, p.author
from rootPosts join posts p on (p.rootParent = rootPosts.id)
order by rootParent desc, mPath;
`

const qSelectPostsPT = `
with rootPosts as (
    select id from posts
    where thread = $1
      and parent = 0
    order by id
    limit $2
    ) select p.id, p.parent, p.message, p.isEdit, p.forum, p.created, p.thread, p.author
from rootPosts join posts p on (p.rootParent = rootPosts.id)
order by mPath;
`

const qSelectPostsPTSinceDesc = `
with rootPosts as (
    select id from posts
    where thread = $1
      and (id < (select rootparent from posts where id = $2))
      and parent = 0
    order by id desc
    limit $3
    ) select p.id, p.parent, p.message, p.isEdit, p.forum, p.created, p.thread, p.author
from rootPosts join posts p on (p.rootParent = rootPosts.id)
order by rootParent desc, mPath;
`

const qSelectPostsPTSince = `
with rootPosts as (
    select id from posts
    where thread = $1
      and (id > (select rootparent from posts where id = $2))
      and parent = 0
    order by id
    limit $3
    ) select p.id, p.parent, p.message, p.isEdit, p.forum, p.created, p.thread, p.author
from rootPosts join posts p on (p.rootParent = rootPosts.id)
order by mPath;
`

const qSelectPostsTDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by mPath desc limit $2`
const qSelectPostsT = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by mPath limit $2`
const qSelectPostsTSinceDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and mPath < (select mPath from posts where id = $2)  order by mPath desc limit $3`
const qSelectPostsTSince = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and mPath > (select mPath from posts where id = $2)  order by mPath limit $3`

const qSelectPostsFDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by id desc limit $2`
const qSelectPostsF = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 order by id limit $2`
const qSelectPostsFSinceDesc = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and id < $2 order by id desc limit $3`
const qSelectPostsFSince = `select id, parent, message, isEdit, forum, created, thread, author from posts where thread = $1 and id > $2 order by id limit $3`

// not prepared
const qUpdateUserFullname = `update users set fullname = $1 where nickname = $2`
const qUpdateUserFullnameAbout = `update users set fullname = $1,about = $2 where nickname = $3`
const qUpdateUserFullnameEmail = `update users set fullname = $1,email = $2 where nickname = $3`
const qUpdateUserFullnameEmailAbout = `update users set fullname = $1,email = $2,about = $3 where nickname = $4`
const qUpdateUserAbout = `update users set about = $1 where nickname = $2`
const qUpdateUserEmail = `update users set email = $1 where nickname = $2`
const qUpdateUserEmailAbout = `update users set email = $1,about = $2 where nickname = $3`
const qUpdatePost = `update posts set isEdit = true, message = $1 where id = $2 returning id, parent, message, isEdit, forum, created, thread, author`

// to prepare
const qInsertUser = `insert into users values ($1, $2, $3, $4)`
const qSelectUserByNickEmail = `select nickname, fullname, about, email from users where nickname = $1 or email = $2`

const qInsertVote = `insert into votes (author, thread, vote) values ($1, $2, $3) on conflict (author, thread) do update set vote = $3`

const qSelectIdForumFromThreadsId = `select id, forum from threads where id = $1::bigint`
const qSelectIdForumFromThreadsSlug = `select id, forum from threads where slug = $1`
const qSelectUsersNickname = `select nickname from users where nickname = $1`
const qInsertForum = `insert into forums (slug, title, owner) values ($1, $2, $3)`
const qCheckForum = `select 1 from forums where slug = $1`
const qSelectSlug = `select slug from forums where slug = $1`
const qUpdateForumPosts = `update forums set posts = posts + $1 where slug = $2`

const qSelectThreadsForumTitle = `select id, title, message, votes, slug, created, forum, author from threads where ((title = $1) and (forum = $2) and (message = $3))`
const qSelectThreadsForumSlug = `select id, title, message, votes, slug, created, forum, author from threads where (slug = $1)`

const qInsertThread = `insert into threads (title, message, author, forum) values ($1, $2, $3, $4) returning id, created`
const qInsertThreadCreated = `insert into threads (title, message, author, forum, created) values ($1, $2, $3, $4, $5) returning id, created`
const qInsertThreadCreatedSlug = `insert into threads (title, message, author, forum, created, slug) values ($1, $2, $3, $4, $5, $6) returning id, created`
const qInsertThreadSlug = `insert into threads (title, message, author, forum, slug) values ($1, $2, $3, $4, $5) returning id, created`

const qSelectThreadBySlug = `select id, title, message, votes, slug, created, forum, author from threads where slug = $1`

type HandlerDB struct {
	pool      *pgx.ConnPool
	bigInsert *pgx.PreparedStatement

	psqSelectForumBySlug       *pgx.PreparedStatement
	psqSelectUserByNick        *pgx.PreparedStatement
	psqSelectThreadById        *pgx.PreparedStatement
	psqSelectPostById          *pgx.PreparedStatement
	psqSelectIdFromThreadsId   *pgx.PreparedStatement
	psqSelectIdFromThreadsSlug *pgx.PreparedStatement

	psqSelectPostsTSinceDesc *pgx.PreparedStatement
	psqSelectPostsTDesc      *pgx.PreparedStatement
	psqSelectPostsTSince     *pgx.PreparedStatement
	psqSelectPostsT          *pgx.PreparedStatement

	psqSelectPostsPTSinceDesc *pgx.PreparedStatement
	psqSelectPostsPTDesc      *pgx.PreparedStatement
	psqSelectPostsPTSince     *pgx.PreparedStatement
	psqSelectPostsPT          *pgx.PreparedStatement

	psqSelectPostsFSinceDesc *pgx.PreparedStatement
	psqSelectPostsFSince     *pgx.PreparedStatement
	psqSelectPostsFDesc      *pgx.PreparedStatement
	psqSelectPostsF          *pgx.PreparedStatement

	psqSelectThreadsCreatedDesc *pgx.PreparedStatement
	psqSelectThreadsDesc        *pgx.PreparedStatement
	psqSelectThreadsCreated     *pgx.PreparedStatement
	psqSelectThreads            *pgx.PreparedStatement

	psqSelectUsersSinceDesc *pgx.PreparedStatement
	psqSelectUsersDesc      *pgx.PreparedStatement
	psqSelectUsersSince     *pgx.PreparedStatement
	psqSelectUsers          *pgx.PreparedStatement

	psqSelectThreadBySlug    *pgx.PreparedStatement
	psqInsertVote            *pgx.PreparedStatement
	psqSelectUserByNickEmail *pgx.PreparedStatement
	psqInsertUser            *pgx.PreparedStatement

	psqInsertThreadCreatedSlug *pgx.PreparedStatement
	psqInsertThreadCreated     *pgx.PreparedStatement
	psqInsertThreadSlug        *pgx.PreparedStatement
	psqInsertThread            *pgx.PreparedStatement

	psqSelectThreadsForumTitle *pgx.PreparedStatement
	psqSelectThreadsForumSlug  *pgx.PreparedStatement

	psqUpdateForumPosts             *pgx.PreparedStatement
	psqSelectIdForumFromThreadsId   *pgx.PreparedStatement
	psqSelectIdForumFromThreadsSlug *pgx.PreparedStatement
	psqSelectUsersNickname          *pgx.PreparedStatement
	psqInsertForum                  *pgx.PreparedStatement

	psqSelectSlug *pgx.PreparedStatement

	// todo atomic cache?
	psqCheckForum *pgx.PreparedStatement
}

func (self *HandlerDB) Connect(config pgx.ConnConfig) (err error) {
	self.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		MaxConnections: 8,
		ConnConfig:     config,
	})

	// todo gsh
	return err
}

const schemaPath = `./base.sql`

func NewHandler() (*HandlerDB, error) {
	config := pgx.ConnConfig{
		Host:     "/run/postgresql",
		Port:     5432,
		User:     "docker",
		Database: "docker",
	}

	handler := &HandlerDB{}

	var err error
	if err = handler.Connect(config); err != nil {
		return nil, err
	}

	if err = handler.CreateTables(schemaPath); err != nil {
		return nil, err
	}

	if err = handler.CreateQueries(); err != nil {
		return nil, err
	}

	return handler, err
}

func (self *HandlerDB) CreateTables(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = self.pool.Exec(string(file))
	if err != nil {
		return err
	}

	return nil
}

func (self *HandlerDB) CreateQueries() error {
	var err error

	self.bigInsert, err = self.pool.Prepare("bigInsert", `insert into posts (parent, message, thread, author, forum) values ($1, $2, $3, $4, $5),($6, $7, $8, $9, $10),($11, $12, $13, $14, $15),($16, $17, $18, $19, $20),($21, $22, $23, $24, $25),($26, $27, $28, $29, $30),($31, $32, $33, $34, $35),($36, $37, $38, $39, $40),($41, $42, $43, $44, $45),($46, $47, $48, $49, $50),($51, $52, $53, $54, $55),($56, $57, $58, $59, $60),($61, $62, $63, $64, $65),($66, $67, $68, $69, $70),($71, $72, $73, $74, $75),($76, $77, $78, $79, $80),($81, $82, $83, $84, $85),($86, $87, $88, $89, $90),($91, $92, $93, $94, $95),($96, $97, $98, $99, $100),($101, $102, $103, $104, $105),($106, $107, $108, $109, $110),($111, $112, $113, $114, $115),($116, $117, $118, $119, $120),($121, $122, $123, $124, $125),($126, $127, $128, $129, $130),($131, $132, $133, $134, $135),($136, $137, $138, $139, $140),($141, $142, $143, $144, $145),($146, $147, $148, $149, $150),($151, $152, $153, $154, $155),($156, $157, $158, $159, $160),($161, $162, $163, $164, $165),($166, $167, $168, $169, $170),($171, $172, $173, $174, $175),($176, $177, $178, $179, $180),($181, $182, $183, $184, $185),($186, $187, $188, $189, $190),($191, $192, $193, $194, $195),($196, $197, $198, $199, $200),($201, $202, $203, $204, $205),($206, $207, $208, $209, $210),($211, $212, $213, $214, $215),($216, $217, $218, $219, $220),($221, $222, $223, $224, $225),($226, $227, $228, $229, $230),($231, $232, $233, $234, $235),($236, $237, $238, $239, $240),($241, $242, $243, $244, $245),($246, $247, $248, $249, $250),($251, $252, $253, $254, $255),($256, $257, $258, $259, $260),($261, $262, $263, $264, $265),($266, $267, $268, $269, $270),($271, $272, $273, $274, $275),($276, $277, $278, $279, $280),($281, $282, $283, $284, $285),($286, $287, $288, $289, $290),($291, $292, $293, $294, $295),($296, $297, $298, $299, $300),($301, $302, $303, $304, $305),($306, $307, $308, $309, $310),($311, $312, $313, $314, $315),($316, $317, $318, $319, $320),($321, $322, $323, $324, $325),($326, $327, $328, $329, $330),($331, $332, $333, $334, $335),($336, $337, $338, $339, $340),($341, $342, $343, $344, $345),($346, $347, $348, $349, $350),($351, $352, $353, $354, $355),($356, $357, $358, $359, $360),($361, $362, $363, $364, $365),($366, $367, $368, $369, $370),($371, $372, $373, $374, $375),($376, $377, $378, $379, $380),($381, $382, $383, $384, $385),($386, $387, $388, $389, $390),($391, $392, $393, $394, $395),($396, $397, $398, $399, $400),($401, $402, $403, $404, $405),($406, $407, $408, $409, $410),($411, $412, $413, $414, $415),($416, $417, $418, $419, $420),($421, $422, $423, $424, $425),($426, $427, $428, $429, $430),($431, $432, $433, $434, $435),($436, $437, $438, $439, $440),($441, $442, $443, $444, $445),($446, $447, $448, $449, $450),($451, $452, $453, $454, $455),($456, $457, $458, $459, $460),($461, $462, $463, $464, $465),($466, $467, $468, $469, $470),($471, $472, $473, $474, $475),($476, $477, $478, $479, $480),($481, $482, $483, $484, $485),($486, $487, $488, $489, $490),($491, $492, $493, $494, $495),($496, $497, $498, $499, $500) returning id, isEdit, created`)

	self.psqSelectForumBySlug, err = self.pool.Prepare("qSelectForumBySlug", qSelectForumBySlug)
	self.psqSelectUserByNick, err = self.pool.Prepare("qSelectUserByNick", qSelectUserByNick)
	self.psqSelectThreadById, err = self.pool.Prepare("qSelectThreadById", qSelectThreadById)
	self.psqSelectPostById, err = self.pool.Prepare("qSelectPostById", qSelectPostById)
	self.psqSelectIdFromThreadsId, err = self.pool.Prepare("qSelectIdFromThreadsId", qSelectIdFromThreadsId)
	self.psqSelectIdFromThreadsSlug, err = self.pool.Prepare("qSelectIdFromThreadsSlug", qSelectIdFromThreadsSlug)

	self.psqSelectPostsPTSinceDesc, err = self.pool.Prepare("qSelectPostsPTSinceDesc", qSelectPostsPTSinceDesc)
	self.psqSelectPostsPTDesc, err = self.pool.Prepare("qSelectPostsPTDesc", qSelectPostsPTDesc)
	self.psqSelectPostsPTSince, err = self.pool.Prepare("qSelectPostsPTSince", qSelectPostsPTSince)
	self.psqSelectPostsPT, err = self.pool.Prepare("qSelectPostsPT", qSelectPostsPT)

	self.psqSelectPostsTSinceDesc, err = self.pool.Prepare("qSelectPostsTSinceDesc", qSelectPostsTSinceDesc)
	self.psqSelectPostsTDesc, err = self.pool.Prepare("qSelectPostsTDesc", qSelectPostsTDesc)
	self.psqSelectPostsTSince, err = self.pool.Prepare("qSelectPostsTSince", qSelectPostsTSince)
	self.psqSelectPostsT, err = self.pool.Prepare("qSelectPostsT", qSelectPostsT)

	self.psqSelectPostsFSinceDesc, err = self.pool.Prepare("qSelectPostsFSinceDesc", qSelectPostsFSinceDesc)
	self.psqSelectPostsFDesc, err = self.pool.Prepare("qSelectPostsFDesc", qSelectPostsFDesc)
	self.psqSelectPostsFSince, err = self.pool.Prepare("qSelectPostsFSince", qSelectPostsFSince)
	self.psqSelectPostsF, err = self.pool.Prepare("qSelectPostsF", qSelectPostsF)

	self.psqSelectThreadsCreatedDesc, err = self.pool.Prepare("qSelectThreadsCreatedDesc", qSelectThreadsCreatedDesc)
	self.psqSelectThreadsDesc, err = self.pool.Prepare("qSelectThreadsDesc", qSelectThreadsDesc)
	self.psqSelectThreadsCreated, err = self.pool.Prepare("qSelectThreadsCreated", qSelectThreadsCreated)
	self.psqSelectThreads, err = self.pool.Prepare("qSelectThreads", qSelectThreads)

	self.psqSelectUsersSinceDesc, err = self.pool.Prepare("qSelectUsersSinceDesc", qSelectUsersSinceDesc)
	self.psqSelectUsersDesc, err = self.pool.Prepare("qSelectUsersDesc", qSelectUsersDesc)
	self.psqSelectUsersSince, err = self.pool.Prepare("qSelectUsersSince", qSelectUsersSince)
	self.psqSelectUsers, err = self.pool.Prepare("qSelectUsers", qSelectUsers)

	// prep
	self.psqSelectThreadBySlug, err = self.pool.Prepare("qSelectThreadBySlug", qSelectThreadBySlug)
	self.psqInsertVote, err = self.pool.Prepare("qInsertVote", qInsertVote)
	self.psqSelectUserByNickEmail, err = self.pool.Prepare("qSelectUserByNickEmail", qSelectUserByNickEmail)
	self.psqInsertUser, err = self.pool.Prepare("qInsertUser", qInsertUser)

	self.psqInsertThreadCreatedSlug, err = self.pool.Prepare("qInsertThreadCreatedSlug", qInsertThreadCreatedSlug)
	self.psqInsertThreadCreated, err = self.pool.Prepare("qInsertThreadCreated", qInsertThreadCreated)
	self.psqInsertThreadSlug, err = self.pool.Prepare("qInsertThreadSlug", qInsertThreadSlug)
	self.psqInsertThread, err = self.pool.Prepare("qInsertThread", qInsertThread)

	self.psqSelectThreadsForumTitle, err = self.pool.Prepare("qSelectThreadsForumTitle", qSelectThreadsForumTitle)
	self.psqSelectThreadsForumSlug, err = self.pool.Prepare("qSelectThreadsForumSlug", qSelectThreadsForumSlug)

	self.psqUpdateForumPosts, err = self.pool.Prepare("qUpdateForumPosts", qUpdateForumPosts)
	self.psqSelectIdForumFromThreadsId, err = self.pool.Prepare("qSelectIdForumFromThreadsId", qSelectIdForumFromThreadsId)
	self.psqSelectIdForumFromThreadsSlug, err = self.pool.Prepare("qSelectIdForumFromThreadsSlug", qSelectIdForumFromThreadsSlug)
	self.psqSelectUsersNickname, err = self.pool.Prepare("qSelectUsersNickname", qSelectUsersNickname)
	self.psqInsertForum, err = self.pool.Prepare("qInsertForum", qInsertForum)

	self.psqSelectSlug, err = self.pool.Prepare("qSelectSlug", qSelectSlug)

	// todo atomic cache?
	self.psqCheckForum, err = self.pool.Prepare("qCheckForum", qCheckForum)

	// todo gsh

	return err
}
