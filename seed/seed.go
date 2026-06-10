package seed

import (
	"encoding/json"
	"log"

	"go-chain/database"
	"go-chain/models"
)

func Seed() {
	db := database.GetDB()

	// Check if data already exists
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count > 0 {
		log.Println("Seed data already exists, skipping")
		return
	}

	// Create categories
	categories := []models.Category{
		{Name: "川菜", Icon: "🌶️"},
		{Name: "粤菜", Icon: "🥟"},
		{Name: "湘菜", Icon: "🔥"},
		{Name: "甜点", Icon: "🍰"},
		{Name: "早餐", Icon: "🌅"},
		{Name: "汤羹", Icon: "🍲"},
		{Name: "素菜", Icon: "🥬"},
		{Name: "面食", Icon: "🍜"},
	}
	db.Create(&categories)

	// Build category lookup
	catMap := make(map[string]uint)
	for _, c := range categories {
		catMap[c.Name] = c.ID
	}

	// Recipe data
	type recipeData struct {
		Name        string
		Category    string
		Ingredients []string
		Steps       []string
		CookTime    string
		Difficulty  string
	}

	recipes := []recipeData{
		// === 川菜 ===
		{
			Name: "麻婆豆腐", Category: "川菜",
			Ingredients: []string{"嫩豆腐 400g", "猪肉末 100g", "豆瓣酱 2勺", "花椒粒 1勺", "蒜末 1勺", "葱花 适量", "淀粉 1勺"},
			Steps:       []string{"豆腐切小块，开水焯2分钟沥干", "锅中热油，下肉末炒至变白", "加豆瓣酱、蒜末炒出红油", "加适量水烧开，放入豆腐轻轻推匀", "淋入水淀粉勾芡，撒花椒粉和葱花出锅"},
			CookTime: "15分钟", Difficulty: "简单",
		},
		{
			Name: "宫保鸡丁", Category: "川菜",
			Ingredients: []string{"鸡胸肉 300g", "花生米 50g", "干辣椒 10个", "花椒 1勺", "葱段 适量", "姜蒜 适量", "醋 2勺", "糖 1勺", "生抽 2勺"},
			Steps:       []string{"鸡胸肉切丁，加料酒、淀粉腌制15分钟", "调汁：醋、糖、生抽、淀粉、水搅匀", "热油小火炒花生米至金黄捞出", "爆香干辣椒和花椒，下鸡丁翻炒至变色", "加入葱姜蒜炒香，倒入调汁翻炒均匀", "最后加入花生米快速翻炒出锅"},
			CookTime: "25分钟", Difficulty: "中等",
		},
		{
			Name: "回锅肉", Category: "川菜",
			Ingredients: []string{"五花肉 400g", "蒜苗 3根", "豆瓣酱 2勺", "甜面酱 1勺", "姜片 3片", "料酒 1勺", "糖 少许"},
			Steps:       []string{"五花肉冷水下锅，加姜片、料酒煮20分钟至八成熟", "捞出放凉切薄片", "蒜苗切斜段", "锅中不放油，下肉片煸炒至卷曲出油", "加豆瓣酱、甜面酱炒香", "放入蒜苗翻炒均匀，加糖提鲜出锅"},
			CookTime: "35分钟", Difficulty: "中等",
		},
		{
			Name: "水煮鱼", Category: "川菜",
			Ingredients: []string{"草鱼 1条", "豆芽 200g", "干辣椒 20g", "花椒 2勺", "豆瓣酱 3勺", "姜蒜 适量", "淀粉 2勺", "料酒 2勺"},
			Steps:       []string{"鱼片成薄片，加料酒、淀粉、蛋清腌制20分钟", "豆芽焯水铺在碗底", "锅中炒香豆瓣酱、姜蒜，加水烧开", "放入鱼片煮至变白，连汤倒入豆芽碗中", "撒上干辣椒段和花椒，淋上热油即可"},
			CookTime: "40分钟", Difficulty: "困难",
		},
		{
			Name: "鱼香肉丝", Category: "川菜",
			Ingredients: []string{"猪里脊 300g", "木耳 50g", "胡萝卜 1根", "青椒 1个", "郫县豆瓣酱 1勺", "醋 2勺", "糖 2勺", "生抽 1勺", "淀粉 1勺"},
			Steps:       []string{"里脊切丝，加料酒、盐、淀粉腌制", "木耳泡发切丝，胡萝卜、青椒切丝", "调鱼香汁：醋、糖、生抽、淀粉、水搅匀", "热油滑熟肉丝盛出", "炒香豆瓣酱，下三丝翻炒", "放入肉丝，倒入鱼香汁翻炒均匀出锅"},
			CookTime: "20分钟", Difficulty: "中等",
		},
		// === 粤菜 ===
		{
			Name: "白切鸡", Category: "粤菜",
			Ingredients: []string{"三黄鸡 1只", "姜 1块", "葱 3根", "料酒 2勺", "冰水 1盆", "蒜 适量", "生抽 3勺", "香油 1勺"},
			Steps:       []string{"鸡处理干净，姜切片，葱打结", "锅中水烧开，放入姜葱料酒", "整鸡放入，大火煮10分钟，关火焖15分钟", "捞出立即放入冰水中浸泡10分钟（皮更脆）", "斩块装盘", "蘸料：蒜末+生抽+香油+少许煮鸡汤拌匀"},
			CookTime: "40分钟", Difficulty: "中等",
		},
		{
			Name: "叉烧肉", Category: "粤菜",
			Ingredients: []string{"梅花肉 500g", "叉烧酱 3勺", "蜂蜜 2勺", "生抽 2勺", "料酒 1勺", "蒜末 适量"},
			Steps:       []string{"梅花肉切厚片，用叉子扎孔便于入味", "叉烧酱+蜂蜜+生抽+料酒+蒜末调成腌料", "肉放入腌料中，冰箱腌制至少4小时", "烤箱预热200°C，烤20分钟", "刷一层蜂蜜，翻面再烤15分钟", "取出稍微放凉，切片装盘"},
			CookTime: "45分钟（不含腌制）", Difficulty: "中等",
		},
		{
			Name: "清蒸鲈鱼", Category: "粤菜",
			Ingredients: []string{"鲈鱼 1条", "姜 1块", "葱 2根", "蒸鱼豉油 3勺", "料酒 1勺", "花生油 2勺"},
			Steps:       []string{"鱼处理干净，两面划几刀", "鱼身抹料酒，放姜片、葱段去腥", "水开后放入蒸锅，大火蒸8-10分钟", "倒掉盘中的腥水", "铺上葱丝姜丝，淋上蒸鱼豉油", "热油浇在葱姜丝上激出香味"},
			CookTime: "15分钟", Difficulty: "简单",
		},
		{
			Name: "干炒牛河", Category: "粤菜",
			Ingredients: []string{"河粉 500g", "牛肉 150g", "豆芽 100g", "韭黄 50g", "洋葱 半个", "生抽 2勺", "老抽 1勺", "糖 少许"},
			Steps:       []string{"牛肉切薄片，加生抽、淀粉腌制", "河粉用温水泡散沥干", "热锅冷油滑熟牛肉盛出", "爆香洋葱，放入豆芽翻炒", "加入河粉，用筷子快速拨散", "加生抽、老抽、糖调色调味", "放入牛肉、韭黄翻炒均匀出锅"},
			CookTime: "15分钟", Difficulty: "中等",
		},
		{
			Name: "煲仔饭", Category: "粤菜",
			Ingredients: []string{"大米 200g", "腊肠 2根", "青菜 几棵", "鸡蛋 1个", "生抽 2勺", "老抽 1勺", "蚝油 1勺", "油 适量"},
			Steps:       []string{"大米提前浸泡30分钟", "砂锅底刷油，放入米和水（没过米一指）", "大火烧开转小火焖至水分收干", "铺上切好的腊肠片，打入鸡蛋", "沿锅边淋一圈油（形成锅巴的关键）", "焖10分钟，放入焯好的青菜", "淋上调好的酱汁，拌匀食用"},
			CookTime: "30分钟", Difficulty: "中等",
		},
		// === 湘菜 ===
		{
			Name: "剁椒鱼头", Category: "湘菜",
			Ingredients: []string{"胖头鱼头 1个", "剁椒 100g", "姜 1块", "蒜 5瓣", "葱花 适量", "料酒 2勺", "蒸鱼豉油 2勺"},
			Steps:       []string{"鱼头处理干净，从中间劈开不切断", "抹上料酒、姜片腌制15分钟", "盘中铺姜片，放上鱼头", "均匀铺上剁椒和蒜末", "水开后大火蒸12分钟", "淋蒸鱼豉油，撒葱花，浇热油"},
			CookTime: "25分钟", Difficulty: "中等",
		},
		{
			Name: "小炒黄牛肉", Category: "湘菜",
			Ingredients: []string{"黄牛肉 300g", "小米椒 10个", "泡椒 5个", "蒜 5瓣", "香菜 1把", "生抽 2勺", "淀粉 1勺"},
			Steps:       []string{"牛肉切薄片，加生抽、淀粉腌制10分钟", "小米椒、泡椒切圈，蒜切末", "热油大火快速滑炒牛肉至变色盛出", "爆香蒜末、辣椒", "倒入牛肉快速翻炒，加生抽调味", "放入香菜段翻炒几下出锅"},
			CookTime: "15分钟", Difficulty: "简单",
		},
		{
			Name: "口味虾", Category: "湘菜",
			Ingredients: []string{"小龙虾 1000g", "干辣椒 20g", "花椒 2勺", "姜蒜 适量", "豆瓣酱 2勺", "啤酒 1罐", "紫苏叶 少许", "糖 1勺"},
			Steps:       []string{"小龙虾刷洗干净，剪去虾须虾线", "热油高温炸至变红捞出", "锅中留底油炒香姜蒜、辣椒、花椒", "加豆瓣酱炒出红油", "放入小龙虾翻炒，倒入啤酒", "加糖、紫苏叶，大火收汁即可"},
			CookTime: "40分钟", Difficulty: "困难",
		},
		// === 甜点 ===
		{
			Name: "芒果糯米饭", Category: "甜点",
			Ingredients: []string{"糯米 200g", "芒果 2个", "椰浆 200ml", "糖 3勺", "盐 少许", "芝麻 少许"},
			Steps:       []string{"糯米浸泡4小时以上或过夜", "蒸锅铺纱布，蒸糯米30分钟至熟", "椰浆加糖、少许盐小火加热搅匀", "蒸好的糯米倒入椰浆中拌匀，焖10分钟", "芒果切片或切块", "糯米盛入盘中，摆上芒果，淋上剩余椰浆，撒芝麻"},
			CookTime: "60分钟（含浸泡）", Difficulty: "简单",
		},
		{
			Name: "双皮奶", Category: "甜点",
			Ingredients: []string{"全脂牛奶 500ml", "蛋清 3个", "糖 40g", "香草精 几滴"},
			Steps:       []string{"牛奶加热至微沸倒入碗中，静置至表面结皮", "用筷子挑开奶皮一角，将牛奶倒出（留少许防粘）", "蛋清加糖打散，与牛奶混合过筛", "混合液沿碗壁倒回有奶皮的碗中", "盖上保鲜膜，中火蒸15分钟", "冷却后冷藏2小时口感更佳"},
			CookTime: "30分钟+冷藏", Difficulty: "中等",
		},
		{
			Name: "蛋挞", Category: "甜点",
			Ingredients: []string{"蛋挞皮 12个", "鸡蛋 2个", "牛奶 100ml", "淡奶油 100ml", "糖 30g", "炼乳 1勺"},
			Steps:       []string{"鸡蛋打散，加入糖搅匀", "加入牛奶、淡奶油、炼乳搅拌均匀", "过筛两次去除蛋筋气泡", "蛋挞皮摆入烤盘，倒入蛋液至八分满", "烤箱预热200°C，烤20-25分钟至表面焦黄"},
			CookTime: "35分钟", Difficulty: "简单",
		},
		// === 早餐 ===
		{
			Name: "鸡蛋灌饼", Category: "早餐",
			Ingredients: []string{"面粉 200g", "鸡蛋 2个", "生菜 几片", "火腿肠 1根", "甜面酱 适量", "葱花 适量"},
			Steps:       []string{"温水和面揉成光滑面团，醒20分钟", "面团分成剂子，擀成薄片", "平底锅刷油，放入面片小火煎", "面皮鼓起来后用筷子挑破，倒入蛋液", "翻面煎至金黄", "刷甜面酱，放生菜、火腿肠卷起即可"},
			CookTime: "20分钟", Difficulty: "简单",
		},
		{
			Name: "皮蛋瘦肉粥", Category: "早餐",
			Ingredients: []string{"大米 150g", "皮蛋 2个", "瘦肉 100g", "姜丝 适量", "葱花 适量", "盐 适量", "香油 少许"},
			Steps:       []string{"大米洗净，加少许油和盐腌制15分钟", "瘦肉切丝或切末", "米加水大火煮开，转小火熬30分钟", "皮蛋切丁，和姜丝一起放入粥中", "继续熬10分钟至粥稠", "加盐调味，撒葱花、滴香油即可"},
			CookTime: "45分钟", Difficulty: "简单",
		},
		{
			Name: "葱油饼", Category: "早餐",
			Ingredients: []string{"面粉 300g", "葱 1把", "盐 适量", "油 适量", "五香粉 少许"},
			Steps:       []string{"开水和面，揉成光滑面团醒30分钟", "葱切葱花", "面团擀成薄片，刷油、撒盐、五香粉和葱花", "卷起后盘成圆形，再擀成饼", "平底锅油热后放入饼，中小火煎至两面金黄", "用铲子轻拍可出层次"},
			CookTime: "30分钟", Difficulty: "中等",
		},
		// === 汤羹 ===
		{
			Name: "番茄牛腩汤", Category: "汤羹",
			Ingredients: []string{"牛腩 500g", "番茄 3个", "土豆 1个", "胡萝卜 1根", "姜片 3片", "番茄酱 2勺", "盐 适量"},
			Steps:       []string{"牛腩切块焯水去血沫", "番茄划十字烫水去皮切块", "锅中炒香姜片，下番茄炒出汤汁", "加番茄酱翻炒，放入牛腩", "加水没过食材，大火烧开转小火炖1.5小时", "加入土豆块、胡萝卜块再炖30分钟", "加盐调味即可"},
			CookTime: "2小时", Difficulty: "中等",
		},
		{
			Name: "玉米排骨汤", Category: "汤羹",
			Ingredients: []string{"排骨 500g", "甜玉米 2根", "胡萝卜 1根", "姜片 3片", "枸杞 少许", "盐 适量"},
			Steps:       []string{"排骨焯水去血沫，洗净", "玉米切段，胡萝卜切滚刀块", "排骨、姜片放入锅中加水大火烧开", "转小火煲1小时", "加入玉米、胡萝卜继续煲30分钟", "加枸杞、盐调味再煮5分钟即可"},
			CookTime: "1.5小时", Difficulty: "简单",
		},
		{
			Name: "冬瓜薏米老鸭汤", Category: "汤羹",
			Ingredients: []string{"老鸭 半只", "冬瓜 500g", "薏米 50g", "姜片 3片", "陈皮 1片", "盐 适量"},
			Steps:       []string{"薏米提前浸泡1小时", "老鸭斩块焯水去血沫", "鸭块、薏米、姜片、陈皮放入锅中加水", "大火烧开转小火煲1.5小时", "冬瓜去皮去瓤切大块，放入继续煲30分钟", "加盐调味即可"},
			CookTime: "2小时", Difficulty: "中等",
		},
		// === 素菜 ===
		{
			Name: "地三鲜", Category: "素菜",
			Ingredients: []string{"茄子 1个", "土豆 2个", "青椒 2个", "蒜末 适量", "生抽 2勺", "糖 1勺", "淀粉 1勺"},
			Steps:       []string{"茄子、土豆切滚刀块，青椒切片", "土豆炸至金黄捞出，茄子炸软捞出", "调汁：生抽、糖、淀粉、水搅匀", "锅中留底油炒香蒜末", "放入所有食材，倒入调汁翻炒均匀", "大火收汁即可出锅"},
			CookTime: "20分钟", Difficulty: "简单",
		},
		{
			Name: "干锅花菜", Category: "素菜",
			Ingredients: []string{"花菜 1颗", "干辣椒 5个", "蒜片 适量", "生抽 2勺", "蚝油 1勺", "糖 少许"},
			Steps:       []string{"花菜掰小朵，盐水浸泡10分钟沥干", "热油爆香蒜片和干辣椒", "放入花菜大火翻炒", "加少许水焖2分钟", "加生抽、蚝油、糖调味", "大火收汁至微微焦香即可"},
			CookTime: "15分钟", Difficulty: "简单",
		},
		{
			Name: "蒜蓉空心菜", Category: "素菜",
			Ingredients: []string{"空心菜 500g", "蒜 6瓣", "盐 适量", "油 适量"},
			Steps:       []string{"空心菜摘洗干净，掐成段", "蒜拍碎切末", "热锅热油，下一半蒜末爆香", "放入空心菜大火快速翻炒", "菜变软后加盐和剩余蒜末", "翻炒均匀立即出锅"},
			CookTime: "5分钟", Difficulty: "简单",
		},
		// === 面食 ===
		{
			Name: "红烧牛肉面", Category: "面食",
			Ingredients: []string{"牛腱子 500g", "面条 300g", "青菜 几棵", "八角 2个", "桂皮 1块", "生抽 3勺", "老抽 1勺", "姜片 3片", "冰糖 10g"},
			Steps:       []string{"牛腱子切大块焯水", "锅中炒糖色，放入牛肉块翻炒上色", "加生抽、老抽、八角、桂皮、姜片", "加水没过牛肉，大火烧开转小火炖1.5小时", "面条煮熟捞出放入碗中", "浇上牛肉和汤，放上焯好的青菜"},
			CookTime: "2小时", Difficulty: "中等",
		},
		{
			Name: "武汉热干面", Category: "面食",
			Ingredients: []string{"碱水面 300g", "芝麻酱 3勺", "生抽 2勺", "醋 1勺", "辣椒油 适量", "葱花 适量", "榨菜末 适量", "香油 1勺"},
			Steps:       []string{"芝麻酱加香油和少许温水调开", "碱水面煮至八成熟捞出沥干", "在案板上淋油抖散面条降温", "吃的时候面条在开水中烫10秒捞出", "加芝麻酱、生抽、醋、辣椒油", "撒葱花、榨菜末拌匀即可"},
			CookTime: "15分钟", Difficulty: "简单",
		},
		{
			Name: "鲜肉馄饨", Category: "面食",
			Ingredients: []string{"馄饨皮 30张", "猪肉末 200g", "虾仁 100g", "姜末 适量", "葱花 适量", "生抽 2勺", "香油 1勺", "紫菜 少许", "蛋皮丝 适量"},
			Steps:       []string{"猪肉末加虾仁泥、姜末、葱花拌匀", "加生抽、香油、少许水搅打上劲", "取馄饨皮包入馅料，捏紧", "水开下馄饨，煮至浮起再煮2分钟", "碗中放紫菜、蛋皮丝、生抽、香油", "舀入馄饨和汤，撒葱花即可"},
			CookTime: "30分钟", Difficulty: "中等",
		},
	}

	// Insert recipes
	for _, r := range recipes {
		ingJSON, _ := json.Marshal(r.Ingredients)
		stepsJSON, _ := json.Marshal(r.Steps)

		catID := catMap[r.Category]
		recipe := models.Recipe{
			Name:        r.Name,
			ImageURL:    "",
			CategoryID:  catID,
			Ingredients: string(ingJSON),
			Steps:       string(stepsJSON),
			CookTime:    r.CookTime,
			Difficulty:  r.Difficulty,
		}
		db.Create(&recipe)
	}

	log.Printf("Seed completed: %d categories, %d recipes", len(categories), len(recipes))
}
