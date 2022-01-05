document.getElementById('get_work').addEventListener('click', () => {
	requestWork()
})
//document.getElementById('gen_doc').addEventListener('click', () => {
//	genDoc()
//})

async function requestWork() {
    const get_work = document.getElementById('get_work')
    get_work.value = "Загружаю..."
	const tasks_container = document.getElementById('tasks_container')
    while (tasks_container.lastChild) {
        tasks_container.removeChild(tasks_container.lastChild);
    }
	const work_tasks = document.getElementById('work_tasks')
	work_tasks.classList.add("hidden")
	const token = document.getElementById('token').value
	const resp = await fetch(`/solution?token=${token}`)
	const data = await resp.json()
	var tasks = data.tasks
	
	const div = document.createElement('div')

	const doc_link = document.getElementById('doc_link')
	doc_link.href = data.link
	work_tasks.classList.remove("hidden")
    get_work.value = "Получить задание"

	for (const task_num in tasks) {
		var task = tasks[task_num]
		const div = document.createElement('div')
		div.classList.add('task')

		const label = document.createElement('label')
		label.classList.add('task_label')
		label.innerHTML = `Задание ${task_num}. ${task.name}`

		const input = document.createElement('input')
		input.classList.add('task_file')
		input.type = 'file'
		input.name = task.task_id
		input.accept = '.c'
		input.classList.add('files')

		const button = document.createElement('button')
		button.classList.add('task_button')
		button.classList.add('send')
		button.innerHTML = 'Проверить'
		button.name = task.task_id
		button.addEventListener('click', send)

		const task_test_cases = document.createElement('textarea')
		task_test_cases.classList.add('task_test_cases')
		task_test_cases.placeholder = `Напиши здесь свои наборы тестовых данных. Пример тестовых наборов:
1;1;
2;3;7;
7;5;1;4;-1;-4;3;`

		const task_text = document.createElement('textarea')
		task_text.classList.add('task_text')
		task_text.readOnly = true

		div.appendChild(label)
		div.appendChild(input)
		div.appendChild(button)
		div.appendChild(task_test_cases)
		div.appendChild(task_text)
		tasks_container.appendChild(div)
	}
	document.getElementById('work_tasks').hidden = false
}

async function send() {
	this.style.disabled = true
	const task_text = this.parentNode.querySelectorAll('.task_text')[0]
	const task_test_cases = this.parentNode.querySelectorAll('.task_test_cases')[0]
	const fileNode = this.parentNode.querySelectorAll('.files')[0]
	const formData = new FormData();
	var file = fileNode.files[0]
	if (!file) {
		task_text.innerHTML = "Файл с решением не приложен!"
		return
	}
	task_text.style = 'height: 1px;'
	task_text.innerHTML = "Решение проверяется..."
    task_text.style = `height: ${task_text.scrollHeight}px;`
	fileNode.value = ""
	formData.append(`source_${this.name}`, file);
	formData.append(`tasks`, this.name);
    test_cases_str = task_test_cases.value
    if (test_cases_str.length > 0 && test_cases_str[test_cases_str.length - 1] !== '\n') {
        test_cases_str += '\n'
    }
	formData.append(`test_cases_${this.name}`, test_cases_str);
	formData.append(`token`, document.getElementById('token').value);
	const resp = await fetch(`/solution`, { method: 'POST', body: formData });
	const data = await resp.json()
	success_msg =
	task_text.style = 'height: 1px;'
	task_text.innerHTML = !data.error ?
        `Задание прошло автотесты!\n\nЗасчитано ошибок: ${data.fail_count}\n\nРезультаты:\n\n${data['result']}` :
        `Задание не прошло автотесты!\n\nЗасчитано ошибок: ${data.fail_count}\n\n${data.error}`;
    task_text.style = `height: ${task_text.scrollHeight}px;`
	this.style.disabled = false
	const tasks_text = document.querySelectorAll('.task_text')
	is_work_complete = true
	for (const text of tasks_text) {
		if (text.innerHTML !== success_msg) {
			is_work_complete = false
			break
		}
	}
	//document.getElementById('report_generation').hidden = !is_work_complete
	//document.getElementById('report_generation_hint').innerHTML = is_work_complete ? 
	//	'Все автотесты пройдены, доступна генерация отчёта' : 
	//	'Генерация отчёта будет доступна после успешного прохождения автотестов'
}

async function genDoc() {
	const report_link = document.getElementById('report_link')
	report_link.innerHTML = ""
	const filesNode = document.querySelectorAll('.files')
	const formData = new FormData();
	var task_ids = []
	for (const file of filesNode) {
		if (!file.files[0]) {
			continue
		}
		formData.append(`source_${file.name}`, file.files[0]);
		task_ids.push(file.name)
	}
	console.log(task_ids.join(','))
	formData.append('tasks', task_ids.join(','))
	formData.append(`token`, document.getElementById('token').value);
	formData.append('name', document.getElementById('name').value)
	formData.append('group', document.getElementById('group').value)
	formData.append('teacher', document.getElementById('teacher').value)
	const resp = await fetch(`/solution?gen_doc=1`, { method: 'POST', body: formData });
	const data = await resp.json()
	report_link.innerHTML = "Сгенерированный отчёт"
	report_link.href = data.link
}

