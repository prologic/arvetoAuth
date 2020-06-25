// Copyright (c) 2020, Arveto Ink. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

namespace User {
	interface User {
		id: string;
		pseudo: string;
		email: string;
		avatar: string;
		level: string;
	}

	export var me: User;
	export var admin: boolean;
	(async function() {
		me = await (await fetch('/me')).json();
		admin = me.level === 'Admin';
	})();

	// List the user in #table
	export async function list() {
		let l: User[] = (await (await fetch('/user/list')).json())
			.filter(u => u.id != me.id);
		(admin ? listAdmin : listSimple)(l)
	}
	async function listSimple(l: User[]) {
		let t = Deskop.table(['', 'ID', 'Pseudo', 'Email', 'Level']);
		l.forEach(u => {
			t.insertAdjacentHTML('beforeend', `<tr>
				<td class="align-middle"><img src="${u.avatar}"
					class="rounded" style="width:3em;"></td>
				<td class="align-middle">${u.id}</td>
				<td class="align-middle">${u.pseudo}</td>
				<td class="align-middle">${u.email}</td>
				<td class="align-middle">${u.level}</td>
			</tr>`);
		});
	}
	async function listAdmin(l: User[]) {
		let t = Deskop.table(['', 'ID', 'Pseudo', 'Email', 'Level', '']);
		l.forEach(u => {
			let tr = $(`<tr>
				<td class="align-middle"><img src="${u.avatar}"
					class="rounded" style="width:3em;"></td>
				<td class="align-middle">${u.id}</td>
				<td class="align-middle">${u.pseudo}</td>
				<td class="align-middle">${u.email}</td>
				<td class="align-middle">${u.level}</td>
				<td class="align-middle">
					<button type=button id=level
						class="btn btn-sm btn-warning">Modifier</button>
					<button type=button id=rm
						class="btn btn-sm btn-danger ml-1">Supprimer</button>
				</td>
			</tr>`);
			t.append(tr);
		});
	}
}
