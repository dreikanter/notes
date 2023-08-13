Page = Data.define(
  :template,
  :layout,
  :local_path,
  :public_path
)

NotePage = Data.define(
  *Page.members,
  :uid,
  :short_uid,
  :slug,
  :tags,
  :published_at,
  :title,
  :body,
  :url
)

TagPage = Data.define(*Page.members, :tag)

RedirectPage = Data.define(*Page.members, :redirect_to)
