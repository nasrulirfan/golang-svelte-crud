<script>
    export let data

    $: ({ employees } = data)
  </script>
  
  <h1>Posts</h1>
  
  <p>Showing {employees.length} posts.</p>
  
  <ul>
  {#each employees as {id , name , email}}
    <li>
      <p>{id}</p>
      <p>{name}</p>
      <p>{email}</p>
    </li>
  {/each}
  </ul>
  