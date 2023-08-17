class Notes::Page
  attr_reader :template, :layout, :local_path, :public_path

  def initialize(attributes = {})
    @template = attributes[:template]
    @layout = attributes[:layout]
    @local_path = attributes[:local_path]
    @public_path = attributes[:public_path]
    @attachments = attributes[:attachments]
  end

  def attachments
    @attachments ||= []
  end
end
